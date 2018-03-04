package deb

import (
	"bytes"
	"crypto/md5"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"time"

	"github.com/goreleaser/nfpm"
	"github.com/pkg/errors"
	"golang.org/x/crypto/openpgp/armor"
	"golang.org/x/crypto/openpgp/clearsign"
	"golang.org/x/crypto/openpgp/packet"
)

type signer struct {
	files []file
	info  nfpm.Info
}

type file struct {
	name   string
	length uint64
	md5Sum string
	shaSum string
}

func (s *signer) add(filename string, data []byte) {
	md5Hasher := md5.New()
	md5Hasher.Write(data)
	md5sum := hex.EncodeToString(md5Hasher.Sum(nil))

	shaHasher := sha1.New()
	shaHasher.Write(data)
	shaSum := hex.EncodeToString(shaHasher.Sum(nil))

	f := file{name: filename, md5Sum: md5sum, shaSum: shaSum, length: uint64(len(data))}
	s.files = append(s.files, f)
}

func (s *signer) sign() ([]byte, error) {
	if s.info.DebSigningKey == "" {
		return nil, nil
	}
	private, uid, err := s.parseKey()
	if err != nil {
		return nil, errors.Wrap(err, "unable to extract information from private key data")
	}
	if private.Encrypted {
		err := private.Decrypt([]byte(s.info.DebSigningKeyPassword))
		if err != nil {
			return nil, errors.Wrap(err, "cannot decrypt private key, check password")
		}
	}
	buffer := bytes.NewBuffer(nil)
	encoder, err := clearsign.Encode(buffer, private, nil)
	if err != nil {
		return nil, errors.Wrap(err, "cannot create signature")
	}
	encoder.Write([]byte(fmt.Sprintf(`Version: 4
Signer: %s
Date: %s
Role: origin
Files:
`, time.Now().Format("Mon Jan 2 15:04:05 2006"), uid.Id)))
	for _, f := range s.files {
		encoder.Write([]byte(fmt.Sprintf("\t%s %s %d %s\n", f.md5Sum, f.shaSum, f.length, f.name)))
	}
	encoder.Close()
	buffer.WriteString("\n")
	return buffer.Bytes(), nil
}

func (s *signer) parseKey() (*packet.PrivateKey, *packet.UserId, error) {
	block, err := armor.Decode(bytes.NewBuffer([]byte(s.info.DebSigningKey)))
	if err != nil {
		return nil, nil, errors.Wrap(err, "cannot decode private key (expecting ascii-armor)")
	}
	reader := packet.NewReader(block.Body)

	var private *packet.PrivateKey
	var uid *packet.UserId

	for {
		pkg, err := reader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, nil, errors.Wrap(err, "cannot read packets of private key")
		}
		switch pkg.(type) {
		case *packet.PrivateKey:
			if private == nil {
				private = pkg.(*packet.PrivateKey)
			}
		case *packet.UserId:
			if uid == nil {
				uid = pkg.(*packet.UserId)
			}
		}
	}
	if private == nil {
		return nil, nil, errors.New("no packet with private key found")
	}
	if uid == nil {
		return nil, nil, errors.New("no packet with user id found")
	}
	return private, uid, nil
}
