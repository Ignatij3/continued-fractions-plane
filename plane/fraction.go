package plane

// fraction represents rational number as a fraction a/b.
type fraction struct {
	a, b uint
}

// gcd returns greatest common divisor of a and b.
func gcd(a, b uint) uint {
	for b != 0 {
		a, b = b, a%b
	}
	return a
}

// getContinuedFraction returns the array which is holding elements of continued fraction of a/b.
func (f fraction) getContinuedFraction() []uint {
	var contFrac []uint = make([]uint, 0)
	for f.b > 0 {
		contFrac = append(contFrac, f.a/f.b)
		f.a, f.b = f.b, f.a%f.b
	}
	return contFrac
}
