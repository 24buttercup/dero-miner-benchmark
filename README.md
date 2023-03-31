# dero-miner-benchmark
requirements: 
`go get -u` all dependencies

run steps:

  1. run the fake pool: go run server.go  
  2. run the miner, connecting to fake pool at the address: 127.0.0.1:14141
  

95% of the code is copied from the official derohe. Please note that this is just a quick mock-up to verify what closed miners doing, approximately benchmark the real hashrate of closed source dero miners. This report was made on 03 Oct 2022, and might be obsolete in the future.

  * The difficulty is fixed at 11111 in the source code.
  * Reported hashrate is the hashrate that shows on the miner console.
  * Theoretical hashrate is the reported hashrate subtracts devfee mining
  * Effective hashrate is calculated from the number of shares as this formula: `FIXED_DIFF * Number_Of_Share / MINING_TIME` or in this case `11111 * Number_Of_Share / MINING_TIME`
  
  * Benchmark time: 1000 seconds (the longer the better)

INTEL BENCHMARK: https://github.com/24buttercup/dero-miner-benchmark/blob/main/INTEL_BENCHMARK.md

AMD BENCHMARK:   https://github.com/24buttercup/dero-miner-benchmark/blob/main/AMD_BENCHMARK.md
