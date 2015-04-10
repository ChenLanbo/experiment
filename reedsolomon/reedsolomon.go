package reedsolomon

type ReedSolomon interface {
    Encode(input []byte) [][]byte
    Decode(input [][]byte) []byte
}
