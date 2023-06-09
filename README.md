# continued-fractions-plane

This program counts amount of times a specific continued fraction's element had appeared. Fractions themselves are made of point's coordinates, so for point with coordinates $(a_x;a_y)$ the resulting fraction would be $\frac{a_x}{a_y}$. Constraints for points are as following:

- They should be no further from origin, than `r`;
- They should be natural numbers;

*Effectively, it means that field of search is restricted to whole points in first quarter of a circle with radius `r` and center in the origin.*

Program takes following command-line arguments:

```txt
-r       - radius of the quarter to go through
-workers - amount of concurrent processes
```

##### ⚠️ If the flags are not present, program does not run
