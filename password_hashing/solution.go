package main

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"golang.org/x/crypto/pbkdf2"
	"golang.org/x/crypto/scrypt"
	"hash"
	"io"
	"net/http"
)

const URL_PROBLEM = "https://hackattic.com/challenges/password_hashing/problem?access_token=986fd83c4079adfd"
const URL_SOLVE = "https://hackattic.com/challenges/password_hashing/solve?access_token=986fd83c4079adfd"

func main() {
	pwResp := getProblem()
	fmt.Println("Problem =======")

	fmt.Printf("Password: %s\n", pwResp.Password)
	fmt.Printf("Salt: %s\n", pwResp.Salt)
	fmt.Printf("Salt Bytes: %v\n", getSaltBytes(pwResp.Salt))
	fmt.Printf("Pbkdf2 Rounds: %d\n", pwResp.Pbkdf2.Rounds)
	fmt.Printf("Pbkdf2 Hash: %s\n", pwResp.Pbkdf2.Hash)
	fmt.Printf("Scrypt N: %d\n", pwResp.Scrypt.N)

	// Verify scrypt
	scryptVerify(pwResp.Scrypt.Control)

	pwSubmit := PasswordSubmit{
		Sha256: sha256Sum(pwResp.Password),
		Hmac:   hmacSha256(pwResp.Password, getSaltBytes(pwResp.Salt)),
		Pbkdf2: pbkdf2Sum(pwResp.Password, getSaltBytes(pwResp.Salt), pwResp.Pbkdf2.Rounds, pwResp.Pbkdf2.Hash),
		Scrypt: scryptSum(pwResp.Password, getSaltBytes(pwResp.Salt), pwResp.Scrypt.N, pwResp.Scrypt.R, pwResp.Scrypt.P, pwResp.Scrypt.Buflen),
	}

	fmt.Println("Solution =======")

	fmt.Printf("Sha256: %s\n", pwSubmit.Sha256)
	fmt.Printf("Hmac: %s\n", pwSubmit.Hmac)
	fmt.Printf("Pbkdf2: %s\n", pwSubmit.Pbkdf2)
	fmt.Printf("Scrypt: %s\n", pwSubmit.Scrypt)

	fmt.Println("Submitting solution...")
	submitSolution(pwSubmit)
}

type PasswordResponse struct {
	Password string `json:"password"`
	Salt     string `json:"salt"`
	Pbkdf2   Pbkdf2 `json:"pbkdf2"`
	Scrypt   Scrypt `json:"scrypt"`
}

type Pbkdf2 struct {
	Rounds int    `json:"rounds"`
	Hash   string `json:"hash"`
}

type Scrypt struct {
	N       int    `json:"N"`
	R       int    `json:"r"`
	P       int    `json:"p"`
	Buflen  int    `json:"buflen"`
	Control string `json:"_control"`
}

type PasswordSubmit struct {
	Sha256 string `json:"sha256"`
	Hmac   string `json:"hmac"`
	Pbkdf2 string `json:"pbkdf2"`
	Scrypt string `json:"scrypt"`
}

func getProblem() PasswordResponse {
	// Make API call to get problem
	res, err := http.Get(URL_PROBLEM)
	if err != nil {
		panic(err)
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		panic(err)
	}

	// Unmarshal response
	var pwResp PasswordResponse
	err = json.Unmarshal(body, &pwResp)
	if err != nil {
		panic(err)
	}
	return pwResp
}

func submitSolution(pwSubmit PasswordSubmit) {
	// Marshal solution
	pwSubmitBytes, err := json.Marshal(pwSubmit)
	if err != nil {
		panic(err)
	}

	// Make API call to submit solution
	res, err := http.Post(URL_SOLVE, "application/json", bytes.NewBuffer(pwSubmitBytes))
	if err != nil {
		panic(err)
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		panic(err)
	}

	fmt.Println(string(body))
}

func getSaltBytes(saltBase64 string) []byte {
	saltBytes, err := base64.StdEncoding.DecodeString(saltBase64)
	if err != nil {
		panic(err)
	}
	return saltBytes
}

func sha256Sum(pw string) string {
	hash := sha256.New()
	hash.Write([]byte(pw))
	return fmt.Sprintf("%x", hash.Sum(nil))
}

func hmacSha256(pw string, secret []byte) string {
	dataBytes := []byte(pw)

	hmacSum := hmac.New(sha256.New, secret)
	hmacSum.Write(dataBytes)
	return fmt.Sprintf("%x", hmacSum.Sum(nil))
}

func pbkdf2Sum(pw string, salt []byte, iter int, hashType string) string {
	var hashFunc func() hash.Hash
	switch hashType {
	case "sha256":
		hashFunc = sha256.New
		break
	case "sha512":
		hashFunc = sha512.New
		break
	default:
		panic("Unsupported hash type")
	}

	dk := pbkdf2.Key([]byte(pw), salt, iter, 32, hashFunc)
	return fmt.Sprintf("%x", dk)
}

func scryptSum(pw string, salt []byte, N, r, p, buflen int) string {
	key, err := scrypt.Key([]byte(pw), salt, N, r, p, buflen)
	if err != nil {
		panic(err)
	}
	return fmt.Sprintf("%x", key)
}

func scryptVerify(control string) {
	password := "rosebud"
	salt := "pepper"
	N := 128
	p := 8
	n := 4

	key, err := scrypt.Key([]byte(password), []byte(salt), N, n, p, 32)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Key: %x\n", key)
	fmt.Printf("Control: %s\n", control)

	if fmt.Sprintf("%x", key) != control {
		panic("Scrypt verification failed")
	}
}
