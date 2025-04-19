package helpers

import (
	"hash/fnv"
	"math/rand"
	"strconv"
	"strings"

	"github.com/speps/go-hashids/v2"
	"golang.org/x/crypto/bcrypt"
)

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func randomString(uuid string) string {
	// Convert UUID to int64 seed
	h := fnv.New64a()
	h.Write([]byte(uuid))
	seed := int64(h.Sum64())

	// Set up rand with the seed
	r := rand.New(rand.NewSource(seed))

	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	b := make([]byte, 12)
	for i := range b {
		b[i] = letters[r.Intn(len(letters))]
	}
	return string(b)
}

func UUIDToFriendlyID(uuid string, name string) string {
	hd := hashids.NewData()
	hd.Salt = name
	hd.MinLength = 12
	h, _ := hashids.NewWithData(hd)

	// Remove hyphens
	cleanUUID := strings.Replace(uuid, "-", "", -1)

	// Convert first 12 chars of UUID to int
	val, _ := strconv.ParseUint(cleanUUID[:12], 16, 64)
	id, _ := h.Encode([]int{int(val)})
	return id
}
