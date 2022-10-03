# dero-miner-benchmark
requirements: 
`go get -u` all dependencies

run steps:

  1. run the fake pool: go run server.go  
  2. run the miner, connecting to fake pool at the address: 127.0.0.1:14141
  

95% of the code is copied from the official derohe. Please note that this is just a quick mock-up to verify what closed miners doing, approximately benchmark the real hashrate of closed source dero miners. This report was made on 03 Oct 2022, and might be obsolete in the future.

  * The difficulty is fixed at 500000 in the source code.
  * Reported hashrate is the hashrate that shows on the miner console.
  * Theoretical hashrate is the reported hashrate subtracts devfee mining
  * Effective hashrate is calculated from the number of shares as this formula: `FIXED_DIFF * Number_Of_Share / MINING_TIME` or in this case `500000 * Number_Of_Share / MINING_TIME`
  * Testing hardware: dual Intel Xeon 8160, 256Gb RAM equally distributed on 6 channels. 
  * Benchmark time: 1000 seconds (the longer the better)

#### miners: ####
- deroplus: https://github.com/Jonutz123/Deroplus-AstroBWTv3/releases
- astrominer: https://github.com/tj8519/astrominer/releases
- stock miner: https://github.com/deroproject/derohe/releases

miner name | reported hashrate | effective hashrate | Theoretical hashrate | dev fee | delta | note |
-----------|-------------------|--------------------|--------------------|---------|-------|------|
deroplus (try1)    | 15.9Khs             |    14.2Khs              |        15.9Khs         |     10%   |     -10%   |   without --show-dev-hashrate. miner mines 10% less than reported hashrate. Dev fee mining running parallely. The delta is ridiculous so I rerun the test again.
deroplus (try2)    | 15.9Khs             |    16.2Khs              |        15.9Khs         |     10%   |     1.8%   |   with --show-dev-hashrate. miner mines 1.8% more than reported hashrate. Probably try1 is just a bad luck. or there is something with this flag "--show-dev-hashrate"
astrominer (try1)  | 25.05Khs            |    26.2Khs              |        22.56Khs        |     10%   |     16%    |   miner mines 16% more than reported hashrate. Dev fee mining is separated by time. The miners got crashed at startup for 2 times, before running smoothly.
astrominer (try2)  | 25.05Khs            |    25.4Khs              |        22.56Khs        |     10%   |     11%    |   miner mines 11% more than reported hashrate. Dev fee mining is separated by time. The miners got crashed at startup for 4 times, before running smoothly. both times, the deltas are high, probably its lucky day.
stock miner        | 13.03Khs            |    13.2Khs              |        13.03Khs        |      0%   |     1.3%   |   miner mines 1.3% more than reported hashrate.

deroplus screenshots:

try 1:

miner screen:

![image](https://user-images.githubusercontent.com/114912206/193548482-c882f1e5-dd8b-4cc6-bf92-1f2034cf0b29.png)

benchmark tool output:

![image](https://user-images.githubusercontent.com/114912206/193548520-78feef72-3839-4ef3-8b1c-61ae3937e7a6.png)

try 2:

miner screen:

![image](https://user-images.githubusercontent.com/114912206/193548559-8d61ea2b-8e25-43e4-88e5-e1e8b9333e18.png)

benchmark tool output:

![image](https://user-images.githubusercontent.com/114912206/193548578-ce718cab-e328-4f3f-b02e-a25566d46a0c.png)


astrominer screenshots:

miner screen:

![image](https://user-images.githubusercontent.com/114912206/193548417-8ef99542-3a42-49fe-a6d2-3f4e59a9b106.png)

benchmark tool output:

![image](https://user-images.githubusercontent.com/114912206/193548324-7c77fcf8-9a5c-4982-9078-eb57cef0cae4.png)
