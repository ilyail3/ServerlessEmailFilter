package comp

import (
	"context"
	"io"
	"google.golang.org/appengine/file"
	"fmt"
	"cloud.google.com/go/storage"
	"golang.org/x/crypto/openpgp"
	"errors"
	"golang.org/x/crypto/openpgp/armor"
)

type PgpVault struct {
	ctx    context.Context
	userId string
}

type encryptionWriter struct {
	openPgpWriter io.WriteCloser
	objectWriter  *storage.Writer
}

func (e *encryptionWriter) Write(p []byte) (int, error) {
	return e.openPgpWriter.Write(p)
}

func (e *encryptionWriter) Close() error {
	err := e.openPgpWriter.Close()

	if err != nil {
		e.objectWriter.Close()
		return err
	}

	err = e.objectWriter.Close()

	if err != nil {
		return err
	}

	return nil
}

var KeyMissing = errors.New("vault: user provided no public key")

func (v *PgpVault) EncryptTo(key string) (io.WriteCloser, error) {
	// Get key
	bucketName, err := file.DefaultBucketName(v.ctx)

	if err != nil {
		return nil, err
	}

	client, err := storage.NewClient(v.ctx)

	if err != nil {
		return nil, err
	}

	bucket := client.Bucket(bucketName)

	publicKey := fmt.Sprintf("keys/%s.gpg", v.userId)

	obj := bucket.Object(publicKey)

	reader, err := obj.NewReader(v.ctx)

	if err == storage.ErrObjectNotExist {
		return nil, KeyMissing
	}

	if err != nil {
		return nil, err
	}

	entityList, err := openpgp.ReadArmoredKeyRing(reader)

	if err != nil {
		return nil, err
	}

	// Output as base64 encoded string
	/*armorWriter,err := armor.Encode(w, "PGP MESSAGE", nil)

	if err != nil {
		http.Error(w, fmt.Sprintf("%v", err), http.StatusInternalServerError)
		return
	}*/

	// Encrypt message using public key

	outputObject := bucket.Object(key)
	w := outputObject.NewWriter(v.ctx)

	writer, err := openpgp.Encrypt(w, entityList, nil, nil, nil)

	if err != nil {
		w.Close()
		return nil, err
	}

	return &encryptionWriter{
		openPgpWriter: writer,
		objectWriter:  w,
	}, nil
}

func (v *PgpVault) ArmorPrint(key string, w io.Writer) error {
	// Get key
	bucketName, err := file.DefaultBucketName(v.ctx)

	if err != nil {
		return err
	}

	client, err := storage.NewClient(v.ctx)

	if err != nil {
		return err
	}

	bucket := client.Bucket(bucketName)

	obj := bucket.Object(key)

	reader, err := obj.NewReader(v.ctx)

	if err != nil {
		return err
	}

	defer reader.Close()

	armorWriter, err := armor.Encode(w, "PGP MESSAGE", map[string]string{"Version": "GnuPG v2"})

	if err != nil {
		return err
	}

	defer armorWriter.Close()

	_, err = io.Copy(armorWriter, reader)

	if err != nil {
		return err
	}

	return nil
}

func NewVault(ctx context.Context, userId string) *PgpVault {
	return &PgpVault{
		ctx:    ctx,
		userId: userId,
	}
}
