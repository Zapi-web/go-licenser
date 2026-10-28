package kgs

import (
	"crypto/rand"
	"fmt"
	"strings"
)

type KGS struct {
	groups   int
	groupLen int
}

const keyAlphabet = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"

func New(groups, groupLen int) *KGS {
	return &KGS{
		groups:   groups,
		groupLen: groupLen,
	}
}

func (k *KGS) GenerateKey() (string, error) {
	b := make([]byte, k.groups*k.groupLen)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("failed to generate key: %w", err)
	}
	var sb strings.Builder
	for i, v := range b {
		if i > 0 && i%k.groupLen == 0 {
			sb.WriteByte('-')
		}
		sb.WriteByte(keyAlphabet[int(v)%len(keyAlphabet)])
	}
	return sb.String(), nil
}
