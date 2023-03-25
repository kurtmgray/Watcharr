package main

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/argon2"
	"gorm.io/gorm"
)

type User struct {
	gorm.Model

	ID       uint   `json:"id"`
	Username string `gorm:"notNull,unique" json:"username" binding:"required"`
	Password string `gorm:"notNnull" json:"password" binding:"required"`
	Watched  []Watched
}

type AuthResponse struct {
	Token string `json:"token"`
}

type ArgonParams struct {
	memory      uint32
	iterations  uint32
	parallelism uint8
	saltLength  uint32
	keyLength   uint32
}

func register(user *User, db *gorm.DB) (AuthResponse, error) {
	println("Registering", user.Username)
	hash, err := hashPassword(user.Password, &ArgonParams{
		memory:      64 * 1024,
		iterations:  3,
		parallelism: 2,
		saltLength:  16,
		keyLength:   32,
	})
	if err != nil {
		log.Fatal(err)
	}

	// Update user obj to replace the plaintext pass with hash
	user.Password = hash

	res := db.Create(&user)
	if res.Error != nil {
		// If error is because unique contraint failed.. user already exists
		if strings.Contains(res.Error.Error(), "UNIQUE") {
			println(err.Error())
			return AuthResponse{}, errors.New("User already exists")
		}
		panic(err)
	}

	// Gorm fills our user obj with the ID from db after insert,
	// just ensure it actually has.
	if user.ID == 0 {
		fmt.Println("user.ID not filled out after registration", user.ID)
		return AuthResponse{}, errors.New("failed to get user id, try login")
	}

	token, err := signJWT(user)
	if err != nil {
		fmt.Println("Failed to sign new jwt:", err)
		return AuthResponse{}, errors.New("failed to get auth token")
	}
	return AuthResponse{Token: token}, nil
}

func login(user *User, db *gorm.DB) (AuthResponse, error) {
	fmt.Println("Logging in", user.Username)
	dbUser := new(User)
	res := db.Where("username = ?", user.Username).Take(&dbUser)
	if res.Error != nil {
		fmt.Println("Failed to select user from database for login:", res.Error)
		return AuthResponse{}, errors.New("User does not exist")
	}

	match, err := compareHash(user.Password, dbUser.Password)
	if err != nil {
		fmt.Println("Failed to compare pass to hash for login:", err)
		return AuthResponse{}, errors.New("failed to login")
	}
	if !match {
		fmt.Println("User failed to provide correct password for login:", match)
		return AuthResponse{}, errors.New("incorrect details")
	}

	token, err := signJWT(dbUser)
	if err != nil {
		fmt.Println("Failed to sign new jwt:", err)
		return AuthResponse{}, errors.New("failed to get auth token")
	}
	return AuthResponse{Token: token}, nil
}

func signJWT(user *User) (token string, err error) {
	// Create new jwt with claim data
	jwt := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"userId":   user.ID,
		"username": user.Username,
	})

	// Sign and get the complete encoded token as a string using the secret
	return jwt.SignedString([]byte(os.Getenv("JWT_SECRET")))
}

func hashPassword(password string, p *ArgonParams) (encodedHash string, err error) {
	salt, err := generateRandomBytes(p.saltLength)
	if err != nil {
		return "", err
	}

	hash := argon2.IDKey([]byte(password), salt, p.iterations, p.memory, p.parallelism, p.keyLength)

	// Base64 encode the salt and hashed password.
	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)

	// Format hash in standard way.
	encodedHash = fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s", argon2.Version, p.memory, p.iterations, p.parallelism, b64Salt, b64Hash)

	return encodedHash, nil
}

func generateRandomBytes(n uint32) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}

	return b, nil
}

func compareHash(password, encodedHash string) (match bool, err error) {
	// Extract the parameters, salt and derived key from the encoded password
	// hash.
	p, salt, hash, err := decodeHash(encodedHash)
	if err != nil {
		return false, err
	}

	// Derive the key from the other password using the same parameters.
	otherHash := argon2.IDKey([]byte(password), salt, p.iterations, p.memory, p.parallelism, p.keyLength)

	// Check that the contents of the hashed passwords are identical. Note
	// that we are using the subtle.ConstantTimeCompare() function for this
	// to help prevent timing attacks.
	if subtle.ConstantTimeCompare(hash, otherHash) == 1 {
		return true, nil
	}
	return false, nil
}

func decodeHash(encodedHash string) (p *ArgonParams, salt, hash []byte, err error) {
	vals := strings.Split(encodedHash, "$")
	if len(vals) != 6 {
		return nil, nil, nil, errors.New("the encoded hash is not in the correct format")
	}

	var version int
	_, err = fmt.Sscanf(vals[2], "v=%d", &version)
	if err != nil {
		return nil, nil, nil, err
	}
	if version != argon2.Version {
		return nil, nil, nil, errors.New("incompatible version of argon2")
	}

	p = &ArgonParams{}
	_, err = fmt.Sscanf(vals[3], "m=%d,t=%d,p=%d", &p.memory, &p.iterations, &p.parallelism)
	if err != nil {
		return nil, nil, nil, err
	}

	salt, err = base64.RawStdEncoding.Strict().DecodeString(vals[4])
	if err != nil {
		return nil, nil, nil, err
	}
	p.saltLength = uint32(len(salt))

	hash, err = base64.RawStdEncoding.Strict().DecodeString(vals[5])
	if err != nil {
		return nil, nil, nil, err
	}
	p.keyLength = uint32(len(hash))

	return p, salt, hash, nil
}
