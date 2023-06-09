goos: windows
goarch: amd64
cpu: AMD Ryzen 5 2600X Six-Core Processor           
BenchmarkInverse/1e2-12         	   30667	     38247 ns/op
BenchmarkInverse/1e3-12         	     326	   3553699 ns/op
BenchmarkInverse/1e4-12         	       3	 352020767 ns/op
BenchmarkFast/1e2-12            	   10000	    107951 ns/op
BenchmarkFast/1e3-12            	      74	  14236151 ns/op
BenchmarkFast/1e4-12            	       1	1809641500 ns/op
BenchmarkFast2/1e2-12           	    8000	    145523 ns/op
BenchmarkFast2/1e3-12           	      36	  32766508 ns/op
BenchmarkFast2/1e4-12           	       1	5509166200 ns/op
BenchmarkMath/1e2-12            	   49791	     24112 ns/op
BenchmarkMath/1e3-12            	     514	   2351336 ns/op
BenchmarkMath/1e4-12            	       5	 230718600 ns/op
PASS
ok  	_/D_/school/ZPD/gauss_distribution/benchmarks/sqrt	22.567s
