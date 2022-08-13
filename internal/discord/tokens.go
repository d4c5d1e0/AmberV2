package discord

import (
	"bufio"
	"math/rand"
	"os"
	"strings"
	"sync"

	"github.com/bytixo/AmberV2/internal/logger"
)

var (
	tokens      []string
	tokensMutex sync.Mutex
)

func init() {
	var err error
	tokens, err = loadTokens()
	if err != nil {
		logger.Error("Failed loading tokens:", err)
		os.Exit(0)
	}
}

func GetTokens() []string {
	return tokens
}
func GetRandomToken() string {
	return tokens[rand.Intn(len(tokens))]
}
func RemoveToken(token string) {
	tokensMutex.Lock()
	defer tokensMutex.Unlock()
	for i, v := range tokens {
		if v == token {
			tokens = append(tokens[:i], tokens[i+1:]...)
		}
	}
	//satomic.AddInt32(&DeadTokens, 1)
	err := rewriteTokens()
	if err != nil {
		logger.Error(err)
	}
}

func loadTokens() ([]string, error) {
	file, err := os.Open("data/tokens.txt")
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}
func rewriteTokens() error {
	file, err := os.OpenFile("data/tokens.txt", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString(strings.Join(tokens, "\n"))
	if err != nil {
		return err
	}
	return nil
}
