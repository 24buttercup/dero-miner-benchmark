package main
 
import (
    "fmt"
    "net/http"

    "sync"
    "encoding/hex"
    "encoding/json"
    "sync/atomic"
    "time"
    "strings"
    "bytes"

    // "bufio"
    // "net"
    // "os"
    // "time"
    // "math/rand"
    "github.com/deroproject/derohe/rpc"
    "github.com/lesismal/llib/std/crypto/tls"
    "github.com/lesismal/nbio/nbhttp"
    "github.com/lesismal/nbio/nbhttp/websocket"
    "github.com/lesismal/nbio"
    "github.com/deroproject/graviton"
    "github.com/deroproject/derohe/globals"
    "runtime"
    "crypto/ecdsa"
    "crypto/elliptic"
    "crypto/rand"
    "crypto/x509"
    "encoding/pem"
    "math/big"
    "github.com/deroproject/derohe/astrobwt/astrobwtv3"
    "github.com/deroproject/derohe/cryptography/crypto"
    randmath "math/rand"
)

var memPool = sync.Pool{
    New: func() interface{} {
        return make([]byte, 16*1024)
    },
}

var HOST_PORT = "0.0.0.0:14141"

var WORK rpc.GetBlockTemplate_Result
var WORK_JSON = "{\"jobid\":\"1664774852132.0.notified\",\"blockhashing_blob\":\"714e2400000f8257deca962a00000000528eef7f98188de81c9d17931e0635e7000000002e81dc49ee9ecd9e6253d583\",\"difficulty\":\"500000\",\"difficultyuint64\":500000,\"height\":1016407,\"prev_hash\":\"deca962a206280d256fe9f0b99751998b9a700d65660ea0fea467cfc6907e461\",\"epochmilli\":0,\"blocks\":0,\"miniblocks\":0,\"rejected\":0,\"lasterror\":\"\",\"status\":\"\"}"

var svr   *nbhttp.Server
type user_session struct {
    blocks        uint64
    miniblocks    uint64
    rejected      uint64
    lasterr       string
    address       rpc.Address
    valid_address bool
    address_sum   [32]byte

    
}

var g_connect_at    int64 = 0
var g_count_print   int = 0

var (
    // bigZero is 0 represented as a big.Int.  It is defined here to avoid
    // the overhead of creating it multiple times.
    bigZero = big.NewInt(0)

    // bigOne is 1 represented as a big.Int.  It is defined here to avoid
    // the overhead of creating it multiple times.
    bigOne = big.NewInt(1)

    // oneLsh256 is 1 shifted left 256 bits.  It is defined here to avoid
    // the overhead of creating it multiple times.
    oneLsh256 = new(big.Int).Lsh(bigOne, 256)

    // enabling this will simulation mode with hard coded difficulty set to 1
    // the variable is knowingly not exported, so no one can tinker with it
    //simulation = false // simulation mode is disabled
)

var client_list_mutex sync.Mutex
var client_list = map[*websocket.Conn]*user_session{}

var miners_count int

var CountMinisAccepted int64 = 0 // total accepted which passed Powtest, chain may still ignore them
var CountMinisRejected int64 // total rejected // note we are only counting rejected as those which didnot pass Pow test
var CountBlocks int64        //  total blocks found as integrator, note that block can still be a orphan
var mini_found_time []int64 // this array contains a epoch timestamp in int64
var rate_lock sync.Mutex

func Accept_new_block(miniblock_blob []byte) bool {
    PoW := astrobwtv3.AstroBWTv3(miniblock_blob)
    block_difficulty := big.NewInt(500000)

    have := HashToBig(PoW)
    want := ConvertIntegerDifficultyToBig(block_difficulty)
    fmt.Printf("Have:%064x\n", have)
    fmt.Printf("Want:%064x\n", want)

    if CheckPowHashBig(PoW, block_difficulty) == true {
        return true
    }
    return false
}

func newUpgrader() *websocket.Upgrader {
    u := websocket.NewUpgrader()

    u.OnMessage(func(c *websocket.Conn, messageType websocket.MessageType, data []byte) {
        // echo
        //c.WriteMessage(messageType, data)

        if messageType != websocket.TextMessage {
            return
        }

        sess := c.Session().(*user_session)

        client_list_mutex.Lock()
        defer client_list_mutex.Unlock()

        var p rpc.SubmitBlock_Params

        if err := json.Unmarshal(data, &p); err != nil {

        }

        mbl_block_data_bytes, err := hex.DecodeString(p.MiniBlockhashing_blob)
        if err != nil {
            //logger.Info("Submitting block could not be decoded")
            sess.lasterr = fmt.Sprintf("Submitted block could not be decoded. err: %s", err)
            return
        }

        var tstamp, extra uint64
        fmt.Sscanf(p.JobID, "%d.%d", &tstamp, &extra)

        sresult := Accept_new_block(mbl_block_data_bytes)

        if sresult {
            fmt.Println("Submitted block accepted")
            atomic.AddInt64(&CountMinisAccepted, 1)
            sess.miniblocks++
        } else {
            fmt.Println("Submitted block rejected")
            atomic.AddInt64(&CountMinisRejected, 1)
            sess.rejected++
        }

    })
    u.OnClose(func(c *websocket.Conn, err error) {
        client_list_mutex.Lock()
        defer client_list_mutex.Unlock()
        delete(client_list, c)
        fmt.Println("------------------------------------------")
        fmt.Println("if you want to benchmark another miner, please restart the program.")
        fmt.Println("------------------------------------------")
    })

    return u
}

func onWebsocket(w http.ResponseWriter, r *http.Request) {
    if !strings.HasPrefix(r.URL.Path, "/ws/") {
        http.NotFound(w, r)
        return
    }
    address := strings.TrimPrefix(r.URL.Path, "/ws/")

    addr, err := globals.ParseValidateAddress(address)
    if err != nil {
        fmt.Fprintf(w, "err: %s\n", err)
        return
    }

    addr_raw := addr.PublicKey.EncodeCompressed()

    upgrader := newUpgrader()
    conn, err := upgrader.Upgrade(w, r, nil)
    if err != nil {
        //panic(err)
        return
    }
    wsConn := conn.(*websocket.Conn)
    wsConn.SetReadDeadline(time.Now().Add(time.Second*30))

    session := user_session{address: *addr, address_sum: graviton.Sum(addr_raw)}
    wsConn.SetSession(&session)

    if g_connect_at == 0 {
        g_connect_at = time.Now().Unix()    
    }
    
    client_list_mutex.Lock()
    defer client_list_mutex.Unlock()
    client_list[wsConn] = &session


}


func StartServer(){
    tlsConfig := &tls.Config{
        Certificates:       []tls.Certificate{generate_random_tls_cert()},
        InsecureSkipVerify: true,
    }
    mux := &http.ServeMux{}
    mux.HandleFunc("/", onWebsocket) // handle everything
    default_address := "0.0.0.0:14141"
    svr = nbhttp.NewServer(nbhttp.Config{
        Name:                    "GETWORK",
        Network:                 "tcp",
        AddrsTLS:                []string{default_address},
        TLSConfig:               tlsConfig,
        Handler:                 mux,
        MaxLoad:                 10 * 1024,
        MaxWriteBufferSize:      32 * 1024,
        ReleaseWebsocketPayload: true,
        NPoller:                 runtime.NumCPU(),
    })
    svr.OnReadBufferAlloc(func(c *nbio.Conn) []byte {
        return memPool.Get().([]byte)
    })
    svr.OnReadBufferFree(func(c *nbio.Conn, b []byte) {
        memPool.Put(b)
    })
    if err := svr.Start(); err != nil {
        fmt.Println(err, "nbio.Start failed.")
        return
    }
    fmt.Println("GETWORK/Websocket server started")
    svr.Wait()
    defer svr.Stop()
}




func main() {
    err := json.Unmarshal([]byte(WORK_JSON), &WORK)
    if err != nil {
        fmt.Println("Error decoding work")
    }

    outstr, _ := json.Marshal(WORK)
    fmt.Println(string(outstr))

    go StartServer()
    for {
        if len(client_list) > 0 {
            SendJob()
        } else {

        }
        time.Sleep(1 * time.Second)        

        g_count_print += 1

        if g_count_print % 20 == 0 {
            total_shares := float64(CountMinisAccepted)
            total_times := float64(time.Now().Unix() - g_connect_at)
            hash_rate := total_shares * 500000 / total_times
            fmt.Println("Hashrate: ", hash_rate, "H/s - Share: ", total_shares, " uptime: ", total_times)
        }
    }
}

// generate default tls cert to encrypt everything
// NOTE: this does NOT protect from individual active man-in-the-middle attacks
func generate_random_tls_cert() tls.Certificate {

    /* RSA can do only 500 exchange per second, we need to be faster
         * reference https://github.com/golang/go/issues/20058
        key, err := rsa.GenerateKey(rand.Reader, 512) // current using minimum size
    if err != nil {
        log.Fatal("Private key cannot be created.", err.Error())
    }

    // Generate a pem block with the private key
    keyPem := pem.EncodeToMemory(&pem.Block{
        Type:  "RSA PRIVATE KEY",
        Bytes: x509.MarshalPKCS1PrivateKey(key),
    })
    */
    // EC256 does roughly 20000 exchanges per second
    key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
    b, err := x509.MarshalECPrivateKey(key)
    if err != nil {
        fmt.Println(err, "Unable to marshal ECDSA private key")
        panic(err)
    }
    // Generate a pem block with the private key
    keyPem := pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: b})

    tml := x509.Certificate{
        SerialNumber: big.NewInt(int64(time.Now().UnixNano())),

        // TODO do we need to add more parameters to make our certificate more authentic
        // and thwart traffic identification as a mass scale

        // you can add any attr that you need
        NotBefore: time.Now().AddDate(0, -1, 0),
        NotAfter:  time.Now().AddDate(1, 0, 0),
        // you have to generate a different serial number each execution
        /*
           Subject: pkix.Name{
               CommonName:   "New Name",
               Organization: []string{"New Org."},
           },
           BasicConstraintsValid: true,   // even basic constraints are not required
        */
    }
    cert, err := x509.CreateCertificate(rand.Reader, &tml, &tml, &key.PublicKey, key)
    if err != nil {
        fmt.Println(err, "Certificate cannot be created.")
        panic(err)
    }

    // Generate a pem block with the certificate
    certPem := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: cert})
    tlsCert, err := tls.X509KeyPair(certPem, keyPem)
    if err != nil {
        fmt.Println(err, "Certificate cannot be loaded.")
        panic(err)
    }
    return tlsCert
}

func CheckPowHashBig(pow_hash crypto.Hash, big_difficulty_integer *big.Int) bool {
    big_pow_hash := HashToBig(pow_hash)

    big_difficulty := ConvertIntegerDifficultyToBig(big_difficulty_integer)
    if big_pow_hash.Cmp(big_difficulty) <= 0 { // if work_pow is less than difficulty
        return true
    }
    return false
}

// HashToBig converts a PoW has into a big.Int that can be used to
// perform math comparisons.
func HashToBig(buf crypto.Hash) *big.Int {
    // A Hash is in little-endian, but the big package wants the bytes in
    // big-endian, so reverse them.
    blen := len(buf) // its hardcoded 32 bytes, so why do len but lets do it
    for i := 0; i < blen/2; i++ {
        buf[i], buf[blen-1-i] = buf[blen-1-i], buf[i]
    }

    return new(big.Int).SetBytes(buf[:])
}

func ConvertIntegerDifficultyToBig(difficultyi *big.Int) *big.Int {

    if difficultyi.Cmp(bigZero) == 0 { // if work_pow is less than difficulty
        panic("difficulty can never be zero")
    }

    return new(big.Int).Div(oneLsh256, difficultyi)
}

func SendJob() {
    defer globals.Recover(1)
    for rk, rv := range client_list {

        func(k *websocket.Conn, v *user_session) {
            defer globals.Recover(2)
            var buf bytes.Buffer
            encoder := json.NewEncoder(&buf)

            var params rpc.GetBlockTemplate_Result
            params = WORK
            // change the work a bit
            rand_str := fmt.Sprintf("%08x",randmath.Intn(2147483647))
            tmp := []byte(params.Blockhashing_blob[:])
            for i := 0; i < 8; i++{
                tmp[12+i] = rand_str[i]
            }
            params.Blockhashing_blob = string(tmp)
            encoder.Encode(params)
            k.SetWriteDeadline(time.Now().Add(2 * time.Second))
            k.WriteMessage(websocket.TextMessage, buf.Bytes())
            buf.Reset()

            k.SetReadDeadline(time.Now().Add(time.Second*30))

        }(rk, rv)

    }

}
