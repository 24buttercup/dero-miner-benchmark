(Report from a discord member)

Testing hardware: dual AMD EPYC 7542
![image](https://user-images.githubusercontent.com/114912206/193732322-a434db7e-576b-44bc-bcf6-124c6610f9a0.png)

If you are not familiar with crypto mining, the delta around +-10% is acceptable in short period (eg: under 3 hours). For longer period (should be 24+ hours), the delta must be below 0.5%.

Interesting note: The user found out that the minimum share that astrominer submits is 65K. He changed the DIFF downto 10k and astrominer continues submitting high DIFF share (65K). So, if you are planning to use astrominer to mine over stratum, don't use fixed diff port that lower than 65K.

miner name | reported hashrate | effective hashrate | Theoretical hashrate | dev fee | delta | note |
-----------|-------------------|--------------------|--------------------|---------|-------|------|
astrominer         | 37.63Khs            |    31.73Khs             |        33.86Khs        |     10%   |     -6%    |
deroplus           | 32.98Khs            |    30.37Khs             |        32.98Khs        |     10%   |     -7.9%  |   
stock miner        | 23.06Khs            |    22.95Khs             |        23.06Khs        |      0%   |     -0.4%  | 


Screenshots:

astrominer:

![image](https://user-images.githubusercontent.com/114912206/193732364-4fbce3a1-0b29-46d0-8f77-c7b7b1b913ed.png)


deroplus:

![image](https://user-images.githubusercontent.com/114912206/193732427-e1755330-4cea-44fb-92d2-43b828277dfe.png)
