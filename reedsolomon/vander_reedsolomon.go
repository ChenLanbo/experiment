package reedsolomon


// Reed Solomon implementation using Vandermonde matrix
type VanderReedSolomon struct {
    // number of data splits
    n int
    // number of code splits
    m int
    // gf for calculation
    gf *GaloisField
    // vandermonde matrix
    vander [][]int
}

func MakeVanderReedSolomon(n, m int) *VanderReedSolomon {
    rs := &VanderReedSolomon{}
    rs.n = n
    rs.m = m
    rs.gf, _ = MakeGaloisField(8)
    rs.vander = make([][]int, rs.m)
    for i := 0; i < rs.m; i++ {
        rs.vander[i] = make([]int, rs.n)
        for j := 1; j <= rs.n; j++ {
            if i == 0 {
                rs.vander[i][j-1] = 1
                continue
            }
            rs.vander[i][j-1] = rs.gf.Multiply(rs.vander[i-1][j-1], j)
        }
    }

    return rs
}

func (rs *VanderReedSolomon) Encode(input []byte) [][]byte {
    output := make([][]byte, rs.n + rs.m)
    buf    := make([]byte, rs.n)
    l := len(input)

    for i := 0; i < l; i += rs.n {
        for j := 0; j < rs.n; j++ {
            if i + j < l {
                buf[j] = input[i + j]
            } else {
                buf[j] = 0
            }
            output[j] = append(output[j], buf[j])
        }
        for j := 0; j < rs.m; j++ {
            c := 0
            for k := 0; k < rs.n; k++ {
                c ^= rs.gf.Multiply(rs.vander[j][k], int(buf[k]))
            }
            output[rs.n + j] = append(output[rs.n + j], byte(c))
        }
    }

    return output
}

func (rs *VanderReedSolomon) Decode(input [][]byte) []byte {
    data := make([][]byte, 0)
    output := make([]byte, 0)

    ma := make([][]int, 0)
    for i := 0; i < rs.n; i++ {
        if len(ma) == rs.n {
            break
        }

        if len(input[i]) != 0 {
            data = append(data, input[i])
            r := make([]int, rs.n)
            r[i] = 1
            ma = append(ma, r)
        }
    }

    for i := 0; i < rs.m; i++ {
        if len(ma) == rs.n {
            break
        }

        if len(input[rs.n + i]) != 0 {
            data = append(data, input[rs.n + i])
            r := make([]int, rs.n)
            copy(r, rs.vander[i])
            ma = append(ma, r)
        }
    }

    if len(ma) != rs.n {
        return output
    }

    ma_i := rs.InverseMatrix(ma)
    buf1 := make([]int, rs.n)
    buf2 := make([]int, rs.n)
    for i := 0; i < len(data[0]); i++ {
        for j := 0; j < rs.n; j++ {
            buf1[j] = int(data[j][i])
            buf2[j] = 0
        }
        for j := 0; j < rs.n; j++ {
            for k := 0; k < rs.n; k++ {
                buf2[j] ^= rs.gf.Multiply(ma_i[j][k], buf1[k])
            }
        }
        for j := 0; j < rs.n; j++ {
            output = append(output, byte(buf2[j]))
        }
    }
    return output
}

func (rs *VanderReedSolomon) InverseMatrix(ma [][]int) [][]int {
    // inverse matrix
    ma_i := make([][]int, len(ma))
    for i := 0; i < len(ma); i++ {
        ma_i[i] = make([]int, len(ma))
    }

    // augmented matrix
    ma_a := make([][]int, len(ma))
    for i := 0; i < len(ma); i++ {
        ma_a[i] = make([]int, len(ma) * 2)
    }
    for i := 0; i < len(ma); i++ {
        for j := 0; j < len(ma); j++ {
            ma_a[i][j] = ma[i][j]
        }
        for j := len(ma); j < 2 * len(ma); j++ {
            ma_a[i][j] = 0
            if i == j - len(ma) {
                ma_a[i][j] = 1
            }
        }
    }

    // gauss-jordan method
    for i := 0; i < len(ma); i++ {
        if ma_a[i][i] == 0 {
            // find a subsequent row to swap
            j := i + 1
            for ; j < len(ma); j++ {
                if ma_a[j][i] != 0 {
                    for k := 0; k < len(ma) * 2; k++ {
                        ma_a[i][k], ma_a[j][k] = ma_a[j][k], ma_a[i][k]
                    }
                }
            }

            if j == len(ma) {
                // singular matrix, error!
            }
        }

        for j := 0; j < len(ma); j++ {
            if j == i {
                continue
            }

            d := rs.gf.Divide(ma_a[j][i], ma_a[i][i])
            for k := 0; k < len(ma) * 2; k++ {
                ma_a[j][k] ^= rs.gf.Multiply(d, ma_a[i][k])
            }
        }
    }

    for i := 0; i < len(ma); i++ {
        if ma_a[i][i] == 1 {
            continue
        }
        d := ma_a[i][i]
        for j := 0; j < len(ma) * 2; j++ {
            ma_a[i][j] = rs.gf.Divide(ma_a[i][j], d)
        }
    }

    for i := 0; i < len(ma); i++ {
        for j := 0; j < len(ma); j++ {
            ma_i[i][j] = ma_a[i][j + len(ma)]
        }
    }

    return ma_i
}

func (rs *VanderReedSolomon) MultiplyMatrices(m1, m2 [][]int) [][]int {
    x1, y1, x2, y2 := len(m1), len(m1[0]), len(m2), len(m2[0])
    if y1 != x2 {
        return nil
    }
    m3 := make([][]int, x1)
    for i, _ := range m3 {
        m3[i] = make([]int, y2)
    }

    for i := 0; i < x1; i++ {
        for j := 0; j < y2; j++ {
            m3[i][j] = 0
            for k := 0; k < y1; k++ {
                m3[i][j] ^= rs.gf.Multiply(m1[i][k], m2[k][j])
            }
        }
    }

    return m3
}
