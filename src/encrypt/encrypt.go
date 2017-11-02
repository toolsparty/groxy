package encrypt

import "github.com/spacemonkeygo/openssl"

type Crypt struct {
	key    []byte
	iv     []byte
	cipher *openssl.Cipher
}

func NewCrypt(key []byte, iv []byte) (*Crypt, error) {
	cipher, err := openssl.GetCipherByName("aes-256-cbc")
	if err != nil {
		return nil, err
	}

	return &Crypt{key, iv, cipher}, nil
}

func (c *Crypt) Encrypt(input []byte) ([]byte, error) {
	ctx, err := openssl.NewEncryptionCipherCtx(c.cipher, nil, c.key, c.iv)
	if err != nil {
		return nil, err
	}

	cipher, err := ctx.EncryptUpdate(input)
	if err != nil {
		return nil, err
	}

	final, err := ctx.EncryptFinal()
	if err != nil {
		return nil, err
	}

	cipher = append(cipher, final...)
	return cipher, nil
}

func (c *Crypt) Decrypt(input []byte) ([]byte, error) {
	ctx, err := openssl.NewDecryptionCipherCtx(c.cipher, nil, c.key, c.iv)
	if err != nil {
		return nil, err
	}

	cipher, err := ctx.DecryptUpdate(input)
	if err != nil {
		return nil, err
	}

	final, err := ctx.DecryptFinal()
	if err != nil {
		return nil, err
	}

	cipher = append(cipher, final...)
	return cipher, nil
}