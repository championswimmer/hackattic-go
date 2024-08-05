package main

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
)

const (
	PROBLEM_URL  = "https://hackattic.com/challenges/help_me_unpack/problem?access_token=986fd83c4079adfd"
	SOLUTION_URL = "https://hackattic.com/challenges/help_me_unpack/solve?access_token=986fd83c4079adfd"
)

func main() {
	problem := getProblem(PROBLEM_URL)

	fmt.Printf("bytes(base64): %s\n", problem.Bytes)

	problemBytes, err := base64.StdEncoding.DecodeString(problem.Bytes)
	if err != nil {
		panic(err)
	}
	fmt.Println("bytes: ", problemBytes)
	fmt.Println("bytes: length = ", len(problemBytes))

	solution := Solution{
		Int:             parseInt32(problemBytes[0:4]),
		Uint:            parseUint32(problemBytes[4:8]),
		Short:           parseShort(problemBytes[8:10]),
		Float:           parseFloat(problemBytes[12:16]),
		Double:          parseDouble(problemBytes[16:24]),
		BigEndianDouble: parseBigEndianDouble(problemBytes[24:32]),
	}
	fmt.Printf("solution: %+v\n", solution)

	submitSolution(SOLUTION_URL, solution)
}

type Problem struct {
	Bytes string `json:"bytes"`
}

type Solution struct {
	Int             int32   `json:"int"`
	Uint            uint32  `json:"uint"`
	Short           int16   `json:"short"`
	Float           float32 `json:"float"`
	Double          float64 `json:"double"`
	BigEndianDouble float64 `json:"big_endian_double"`
}

func getProblem(url string) Problem {
	resp, err := http.Get(url)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	var problem Problem
	err = json.NewDecoder(resp.Body).Decode(&problem)
	if err != nil {
		panic(err)
	}
	return problem
}

func submitSolution(url string, solution Solution) {
	body, err := json.Marshal(solution)
	if err != nil {
		panic(err)
	}

	resp, err := http.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	fmt.Println("response Status:", resp.Status)
	fmt.Println("response Body:", string(respBytes))

}

// take 4 bytes and return int32
func parseInt32(bytes []byte) int32 {
	return int32(bytes[3])<<24 | int32(bytes[2])<<16 | int32(bytes[1])<<8 | int32(bytes[0])
}

func parseUint32(bytes []byte) uint32 {
	return uint32(bytes[3])<<24 | uint32(bytes[2])<<16 | uint32(bytes[1])<<8 | uint32(bytes[0])
}

func parseShort(bytes []byte) int16 {
	return int16(bytes[1])<<8 | int16(bytes[0])
}

func parseFloat(bytes []byte) float32 {
	bits := binary.LittleEndian.Uint32(bytes)
	return math.Float32frombits(bits)
}

func parseDouble(bytes []byte) float64 {
	bits := binary.LittleEndian.Uint64(bytes)
	return math.Float64frombits(bits)
}

func parseBigEndianDouble(bytes []byte) float64 {
	bits := binary.BigEndian.Uint64(bytes)
	return math.Float64frombits(bits)
}
