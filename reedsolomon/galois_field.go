package reedsolomon

import "fmt"
import "errors"

// A Galois field GF(n) is a set of n elements closed under addition and
// multiplication, for which every element has an additive and multiplicative
// inverse (except for the 0 element which has no multiplicative inverse).
//
// Galois field GF(2^w): work with polynomials of x whose coefficients are in
// GF(2).
//
// Construction of GF(2^w):
// 1. Find q(x) which is a primitive polynomial of degree w and whose 
//    coefficients are in GF(2).
// 2. Start with elements 0, 1, x, and then continue to enumerate the elements
//    by multiplying the last element by x and taking the result modulo q(x)
//    if it has a degree >=￼w. The last element multiplied by x mod q(x) = 1.
// 3. Map each element to a binary work b of size whose ith bit corresponds to
//    the coefficient of x^i in the polynomial element.
//
// Example of the construction of GF(2^3):
//
// q(x) = x^3 + x + 1
//
// 0      0              000 -> 0
// x^0    1              001 -> 1
// x^1    x              010 -> 2
// x^2    x^2            100 -> 4
// x^3    x + 1          011 -> 3
// x^4    x^2 + x        110 -> 6
// x^5    x^2 + x + 1    111 -> 7
// x^6    x^2 + 1        101 -> 5
// x^7    1 <-- end
//
// Then:
// glog(1) = 0, gilog(0) = 1
// glog(2) = 1, gilog(1) = 2
// glog(3) = 3, gilog(3) = 3
// glog(4) = 2, gilog(2) = 4
// glog(6) = 4, gilog(4) = 6
// glog(5) = 6, gilog(6) = 5
// glog(7) = 5, gilog(5) = 7

// Supported galois field
const (
    prim_poly_2 int = 7
    prim_poly_3 int = 11
    prim_poly_4 int = 19
    prim_poly_8 int = 285
)

type GaloisField struct {

    // Word size
    W int

    // The table that maps the index (1 to 2^W - 1) to its logarithm.
    Log  []int

    // The table that maps the index (0 to 2^W - 2) to its inverse
    // logarithm
    Ilog []int
}

func (gf *GaloisField) Multiply(a int, b int) int {
    if a == 0 || b == 0 {
        return 0
    }

    log_sum := gf.Log[a] + gf.Log[b]
    if log_sum >= (1 << uint32(gf.W)) - 1 {
        log_sum -= (1 << uint32(gf.W)) - 1
    }
    return gf.Ilog[log_sum]
}

func (gf *GaloisField) Divide(a int, b int) (int, error) {
    if a == 0 {
        return 0, nil
    }
    if b == 0 {
        return 0, errors.New("0")
    }

    log_sub := gf.Log[a] - gf.Log[b]
    if log_sub < 0 {
        log_sub += (1 << uint32(gf.W)) - 1
    }
    return gf.Ilog[log_sub], nil
}

func (gf *GaloisField) PrintInfo() {
    for i := 0; i < (1 << uint32(gf.W)); i++ {
        fmt.Println(i, "log", gf.Log[i], "ilog", gf.Ilog[i])
    }
}

func MakeGaloisField(w int) (*GaloisField, error) {
    var prim_poly int

    switch w {
    case 2:
        prim_poly = prim_poly_2
    case 3:
        prim_poly = prim_poly_3
    case 4:
        prim_poly = prim_poly_4
    case 8:
        prim_poly = prim_poly_8
    default:
        return nil, errors.New("Invalid GaloisField")
    }

    gf := &GaloisField{}
    gf.W = w
    gf.Log = make([]int,  1 << uint32(w))
    gf.Ilog = make([]int, 1 << uint32(w))

    var b int = 1
    var log int = 0
    for ; log < (1 << uint32(w)) - 1; log++ {
        gf.Log[b] = log
        gf.Ilog[log] = b

        b <<= 1
        if (b & (1 << uint32(w))) != 0 {
            b ^= prim_poly
        }
    }

    return gf, nil
}

