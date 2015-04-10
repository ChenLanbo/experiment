package reedsolomon

import "testing"

func TestInverseMatrix(t *testing.T) {
    rs := MakeVanderReedSolomon(4, 2)

    ma := make([][]int, 4)
    for i := 0; i < 4; i++ {
        ma[i] = make([]int, 4)
    }

    // identity matrix
    for i := 0; i < 4; i++ {
        ma[i][i] = 1
    }

    t.Log("Original matrix", ma)
    ma_i := rs.InverseMatrix(ma)
    t.Log("Inverse matrix", ma_i)
    t.Log("Verify", rs.MultiplyMatrices(ma, ma_i))

    // some Vandermonde matrices
    // 1 0 0 0
    // 0 1 0 0
    // 0 0 1 0
    // 1 2 3 4 
    for i, _ := range ma {
        for j, _ := range ma[i] {
            ma[i][j] = 0
        }
    }
    ma[0][0], ma[1][1], ma[2][2] = 1, 1, 1
    for i, _ := range ma[3] {
        ma[3][i] = i + 1
    }
    t.Log("Original matrix", ma)
    ma_i = rs.InverseMatrix(ma)
    t.Log("Inverse matrix", ma_i)
    t.Log("Verify", rs.MultiplyMatrices(ma, ma_i))

    // 1 0 0 0
    // 0 0 1 0
    // 1 1 1 1
    // 1 2 3 4
    for i, _ := range ma {
        for j, _ := range ma[i] {
            ma[i][j] = 0
        }
    }
    ma[0][0], ma[1][2] = 1, 1
    for i, _ := range ma[2] {
        ma[2][i] = 1
    }
    for i, _ := range ma[3] {
        ma[3][i] = i + 1
    }
    t.Log("Original matrix", ma)
    ma_i = rs.InverseMatrix(ma)
    t.Log("Inverse matrix", ma_i)
    t.Log("Verify", rs.MultiplyMatrices(ma, ma_i))
}

