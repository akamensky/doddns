package utils

import (
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"strings"
)

func ParseAndValidateIPAddress(s string) (net.IP, error) {
	ip := net.ParseIP(s)
	if ip == nil {
		return nil, fmt.Errorf("bad IP %s", s)
	}

	// TODO: Actually validate if it is good IP

	return ip, nil
}

func IsIPv4(ip net.IP) bool {
	return ip.To4() != nil
}

func ReadTokenFromFile(fp string) (string, error) {
	fd, err := os.Open(fp)
	if err != nil {
		return "", err
	}
	defer func(fd *os.File) {
		_ = fd.Close()
	}(fd)

	b, err := io.ReadAll(fd)
	if err != nil {
		return "", err
	}
	token := string(b)
	if token == "" {
		return "", fmt.Errorf("token file is empty")
	}

	return token, nil
}

func GetEnvDefaultInt(name string, value int) int {
	env := os.Getenv(name)
	if env == "" {
		return value
	}

	res, err := strconv.Atoi(env)
	if err != nil {
		panic(err)
	}
	return res
}

func GetEnvDefaultStringList(name string, value []string) []string {
	env := os.Getenv(name)
	if env == "" {
		return value
	}

	result := make([]string, 0)
	parts := strings.Split(env, " ")
	for _, part := range parts {
		result = append(result, strings.TrimSpace(part))
	}
	return result
}
