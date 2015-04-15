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

func TestEncodeDecode(t *testing.T) {
    rs1 := MakeVanderReedSolomon(4, 2)
    rs2 := MakeVanderReedSolomon(6, 3)

    // test rs=4.2
    data1 := make([]byte, 12)
    for i, _ := range data1 {
        data1[i] = byte(i + 1)
    }

    encode1 := rs1.Encode(data1)
    t.Log("Encoded:", encode1)

    encode1[0] = make([]byte, 0)
    encode1[3] = make([]byte, 0)
    data2 := rs1.Decode(encode1)
    t.Log("Decoded:", data2)

    // test rs=6.3
    data3 := make([]byte, 24)
    for i, _ := range data3 {
        data3[i] = byte(i + 1)
    }

    encode2 := rs2.Encode(data3)
    t.Log("Encoded:", encode2)

    encode2[3] = make([]byte, 0)
    encode2[5] = make([]byte, 0)
    encode2[7] = make([]byte, 0)
    data4 := rs2.Decode(encode2)
    t.Log("Decoded:", data4)
}
