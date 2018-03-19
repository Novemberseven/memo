package db

import (
	"git.jasonc.me/main/bitcoin/wallet"
	"git.jasonc.me/main/cryptography"
	"github.com/jchavannes/jgo/jerr"
	"time"
)

const KeyTable = "Key"

type Key struct {
	Id        uint `gorm:"primary_key"`
	Name      string
	UserId    uint
	Value     []byte
	PublicKey []byte
	MaxCheck  uint // maximum block height checked for transactions
	MinCheck  uint // minimum block height checked for transactions
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (k *Key) Save() error {
	result := save(k)
	if result.Error != nil {
		return jerr.Get("error saving key", result.Error)
	}
	return nil
}

func (k Key) GetPrivateKey(password string) (*wallet.PrivateKey, error) {
	key, err := cryptography.GenerateKeyFromPassword(password)
	if err != nil {
		return nil, jerr.Get("error generating key from password", err)
	}
	decrypted, err := cryptography.Decrypt(k.Value, key)
	if err != nil {
		return nil, jerr.Get("failed to decrypt", err)
	}
	privateKey := wallet.PrivateKey{
		Secret: decrypted,
	}
	pubKey := privateKey.GetPublicKey().GetSerializedString()
	if pubKey != k.GetPublicKey().GetSerializedString() {
		return nil, jerr.New("error decrypting, public key doesn't match")
	}
	return &privateKey, nil
}

func (k Key) GetPublicKey() wallet.PublicKey {
	return wallet.GetPublicKey(k.PublicKey)
}

func (k Key) GetAddress() wallet.Address {
	return k.GetPublicKey().GetAddress()
}

func (k Key) Delete() error {
	result := remove(&k)
	if result.Error != nil {
		return jerr.Get("error deleting key", result.Error)
	}
	return nil
}

func GenerateKey(name string, password string, userId uint) (*Key, error) {
	key, err := cryptography.GenerateKeyFromPassword(password)
	if err != nil {
		return nil, jerr.Get("error generating key from password", err)
	}
	privateKey := wallet.GeneratePrivateKey()
	return createKey(name, privateKey, key, userId)
}

func ImportKey(name string, password string, wif string, userId uint) (*Key, error) {
	key, err := cryptography.GenerateKeyFromPassword(password)
	if err != nil {
		return nil, jerr.Get("error generating key from password", err)
	}
	privateKey, err := wallet.ImportPrivateKey(wif)
	if err != nil {
		return nil, jerr.Get("error importing key from wif", err)
	}
	return createKey(name, privateKey, key, userId)
}

func createKey(name string, privateKey wallet.PrivateKey, key []byte, userId uint) (*Key, error) {
	encryptedSecret, err := cryptography.Encrypt(privateKey.Secret, key)
	if err != nil {
		return nil, jerr.Get("failed to encrypt", err)
	}
	var dbPrivateKey = &Key{
		Name:      name,
		UserId:    userId,
		Value:     encryptedSecret,
		PublicKey: privateKey.GetPublicKey().GetSerialized(),
	}
	result := save(dbPrivateKey)
	if result.Error != nil {
		return nil, jerr.Get("error saving key", result.Error)
	}
	return dbPrivateKey, nil
}

func GetKey(id uint, userId uint) (*Key, error) {
	var privateKey Key
	err := find(&privateKey, Key{
		Id:     id,
		UserId: userId,
	})
	if err != nil {
		return nil, err
	}
	return &privateKey, nil
}

func GetKeysForUser(userId uint) ([]*Key, error) {
	var privateKeys []*Key
	err := find(&privateKeys, Key{
		UserId: userId,
	})
	if err != nil {
		return nil, err
	}
	return privateKeys, nil
}

func GetAllKeys() ([]*Key, error) {
	var privateKeys []*Key
	err := find(&privateKeys, Key{})
	if err != nil {
		return nil, err
	}
	return privateKeys, nil
}
