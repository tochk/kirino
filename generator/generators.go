package generator

import (
	"crypto/sha256"
	"encoding/hex"
	"math/rand"
	"os"
	"strconv"
	"time"

	"github.com/tochk/kirino/templates/html"
)

type WifiUser = html.WifiUser
type Domain = html.Domain
type Mail = html.Mail

func generateHash(word string) string {
	hasher := sha256.New()
	s1 := rand.NewSource(time.Now().UnixNano())
	r1 := rand.New(s1)
	hashStr := word + strconv.Itoa(r1.Intn(1000000))
	hasher.Write([]byte(hashStr))
	hashedStr := hex.EncodeToString(hasher.Sum(nil))
	if file, err := os.Open("userFiles\\" + hashStr + ".tex"); err == nil {
		file.Close()
		hashedStr = generateHash(hashedStr)
	}
	return hashedStr
}
