package bitwarden

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/meerkat-manor/salainen"
	"github.com/meerkat-manor/salainen/extensions/bitwarden/gen"
	"github.com/oapi-codegen/runtime/types"
)

type f struct {
	ApiUrl      string
	IdentityURL string
	AccessToken string
	ProductName string // Name couldbe VaultWarden, BitWarden or compatible
}

// Bitwarden structures from responses

type FolderList struct {
	Success bool `json:"success,omitempty"`
	Data    struct {
		Object string `json:"object,omitempty"`
		Data   []struct {
			Object string `json:"object,omitempty"`
			Id     string `json:"id,omitempty"`
			Name   string `json:"name,omitempty"`
		} `json:"data,omitempty"`
	} `json:"data,omitempty"`
}

type ItemList struct {
	Success bool `json:"success,omitempty"`
	Data    struct {
		Object string             `json:"object,omitempty"`
		Data   []gen.ItemTemplate `json:"data,omitempty"`
	} `json:"data,omitempty"`
	RevisionDate string `json:"revisionDate,omitempty"`
	DeleteDate   string `json:"deleteDate,omitempty"`
}

type ItemPassword struct {
	Success bool `json:"success,omitempty"`
	Data    struct {
		Object string `json:"object,omitempty"`
		Data   string `json:"data,omitempty"`
	} `json:"data,omitempty"`
}

const providerName = "bitwarden"

func (sl *f) Init(custom interface{}) error {

	sl.ProductName = "Bitwarden"

	if custom != nil {
		if settings, ok := custom.(map[string]string); ok {

			value, exists := settings["ApiUrl"]
			if exists && value != "" {
				sl.ApiUrl = value
			}

			value, exists = settings["IdentityURL"]
			if exists && value != "" {
				sl.IdentityURL = value
			}

			value, exists = settings["AccessToken"]
			if exists && value != "" {
				// NOTE that the value will be unpacked at usage not now
				sl.AccessToken = value
			}
		}
	}
	return nil
}

func (sl *f) Put(path, val string) error {
	parts := strings.SplitN(path, "|", 2)
	if len(parts) > 1 {

	} else {

	}

	return nil
}

func (sl *f) Get(path string) (string, error) {

	getData := ""

	httpClient := &http.Client{}
	server := sl.ApiUrl

	if sl.AccessToken == "" {
		return "", fmt.Errorf("no access token for %s provided", sl.ProductName)
	}
	if strings.HasPrefix(sl.AccessToken, "bitwarden:") {
		return "", fmt.Errorf("error fetching %s access password with looping detected", sl.ProductName)
	}

	password, errS := salainen.Get(sl.AccessToken)
	if errS != nil || password == "" {
		return "", fmt.Errorf("error fetching %s access password.  More information: %v", sl.ProductName, errS)
	}

	unlocked := false

	client, errC := gen.NewClientWithResponses(server, gen.WithHTTPClient(httpClient))
	if errC != nil {
		return "", fmt.Errorf("error making %s client. Error: %v", sl.ProductName, errC)
	}

	// Check status
	resp, err := client.GetStatus(context.Background())
	if err != nil {
		return "", fmt.Errorf("error with status check on %s. Error: %v", sl.ProductName, err)
	} else {
		defer resp.Body.Close()
		builder := new(strings.Builder)
		_, err := io.Copy(builder, resp.Body)
		if err != nil {
			fmt.Printf("error with reading %s response body: %v", sl.ProductName, err)
		} else {
			data := builder.String()

			var status gen.Status
			err = json.Unmarshal([]byte(data), &status)
			if err != nil {
				fmt.Printf("error during Unmarshal(): %s", err)
			}
			if *status.Data.Template.Status == gen.Unlocked {
				unlocked = true
			}
		}
	}

	if !unlocked {
		unlock := gen.PostUnlockJSONRequestBody{
			Password: &password,
		}

		respU, errU := client.PostUnlock(context.Background(), unlock)
		if errU != nil {
			return "", fmt.Errorf("error unlocking %s. Error: %v", sl.ProductName, errU)
		} else {
			if respU.StatusCode != 200 {
				return "", fmt.Errorf("error unlocking %s with status code: %d", sl.ProductName, respU.StatusCode)
				/*
					// Is this needed TODO
					if respU.ContentLength > 0 {
						defer respU.Body.Close()
						builder := new(strings.Builder)
						_, err := io.Copy(builder, respU.Body)
						if err != nil {
							fmt.Printf("error with reading %s response body: %v", sl.ProductName, err)
						} else {
							//data := builder.String()
							//fmt.Printf("DEBUG BW get body: %s\n", data)
						}
					}
				*/
			} else {
				unlocked = true
			}
		}
	}

	if unlocked {

		id, errUI := uuid.Parse(path)
		if errUI == nil {
			getData, err = sl.getItemByUUID(client, id)
			if err != nil {
				return "", fmt.Errorf("error fetching %s item (UUID). Error: %v", sl.ProductName, err)
			}
		} else {

			parts := strings.Split(path, "/")
			if len(parts) < 2 {

			} else {
				parent := parts[0]
				for ix := 1; ix < (len(parts) - 1); ix++ {
					parent = parent + "/" + parts[ix]
				}
				folderId, err := sl.getFolders(client, parent)
				if err != nil {
					return "", fmt.Errorf("error fetching %s folders (path). Error: %v", sl.ProductName, err)
				}
				getData, err = sl.getItemByFolder(client, folderId, &parts[len(parts)-1])
				if err != nil {
					return "", fmt.Errorf("error fetching %s item (path). Error: %v", sl.ProductName, err)
				}
			}

		}

	}

	return getData, nil
}

func (sl *f) Help() {
	fmt.Printf("Bitwarden help\n\n")
	fmt.Printf("Bitwarden or any API compatible password manager can be used\n")
	fmt.Printf("as a secret provider by using the prefix 'bitwarden:' followed\n")
	fmt.Printf("by the path key to the secret or the UUID for the item\n")
	fmt.Printf("\n")
	fmt.Printf("The pre-requisite is that you are running 'bw serve' locally\n")
	fmt.Printf("and have configured the settings for salainen.\n")
	fmt.Printf("Only one Bitwarden source can be defined , the value of which\n")
	fmt.Printf("in the 'ApiURL' and 'IdentityURL' configuration settings.\n")
	fmt.Printf("The master password is in the configuration under 'AccessToken'.\n")
	fmt.Printf("The access token is processed as a 'salainen` value so.\n")
	fmt.Printf("define it using the format '<provider>:<key>' where you could\n")
	fmt.Printf("use for example 'plain:masterpassword' or 'keyring:secretkey \n")
	fmt.Printf("You cannot use 'bitwarden:password' as that would cause an infinite loop.\n")
	fmt.Printf("The provider is only available on platforms supported by Bitwarden bw\n")
	fmt.Printf("Compatible password managers: Vaultwarden\n")
	fmt.Printf("\n")
	fmt.Printf("For more information please see %s/extensions/bitwarden/ \n", salainen.SourceForgeURL)
}

func (sl *f) getFolders(client *gen.ClientWithResponses, folder string) (*types.UUID, error) {

	parms := gen.GetListObjectFoldersParams{
		Search: &folder,
	}

	resp, err := client.GetListObjectFolders(context.Background(), &parms)
	if err != nil {
		return nil, fmt.Errorf("error fetching %s folders. Error: %v", sl.ProductName, err)
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("error fetching %s folders with status code: %d", sl.ProductName, resp.StatusCode)
	}
	if resp.ContentLength > 0 {
		defer resp.Body.Close()
		builder := new(strings.Builder)
		_, err := io.Copy(builder, resp.Body)
		if err != nil {
			return nil, fmt.Errorf("error with reading %s response body: %v", sl.ProductName, err)
		} else {
			data := builder.String()
			var folderList FolderList
			err = json.Unmarshal([]byte(data), &folderList)
			if err != nil {
				//fmt.Printf("error during Unmarshal(): %s", err)
			}

			if folderList.Success && len(folderList.Data.Data) > 0 {
				for _, item := range folderList.Data.Data {
					if item.Name == folder {
						id, err := uuid.Parse(item.Id)
						if err != nil {
							return nil, err
						}
						return &id, nil
					}
				}
			}
		}
	}

	return nil, fmt.Errorf("error when listing folders")
}

func (sl *f) getItemByFolder(client *gen.ClientWithResponses, folderId *types.UUID, itemName *string) (string, error) {

	// fmt.Printf("DEBUG item by folder/name: %v %s\n", folderId, *itemName)

	parms := gen.GetListObjectItemsParams{
		Folderid: folderId,
	}

	resp, err := client.GetListObjectItems(context.Background(), &parms)
	if err != nil {
		return "", fmt.Errorf("error fetching %s item. Error: %v", sl.ProductName, err)
	}
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("error fetching %s item with status code: %d", sl.ProductName, resp.StatusCode)
	}
	if resp.ContentLength > 0 {
		defer resp.Body.Close()
		builder := new(strings.Builder)
		_, err := io.Copy(builder, resp.Body)
		if err != nil {
			fmt.Printf("error with reading %s response body: %v\n", sl.ProductName, err)
		} else {
			data := builder.String()

			var itemList ItemList
			err = json.Unmarshal([]byte(data), &itemList)
			if err != nil {
				// fmt.Printf("error during Unmarshal(): %s", err)
			}

			if itemList.Success && len(itemList.Data.Data) > 0 {
				for _, item := range itemList.Data.Data {
					if *item.Name == *itemName || (item.Login != nil && *item.Login.Username == *itemName) {
						if item.Login.Password == nil {
							return "", fmt.Errorf("no password recorded")
						}
						return *item.Login.Password, nil
					}
				}
			}
		}
	}

	return "", fmt.Errorf("not matching record located")
}

func (sl *f) getItemByUUID(client *gen.ClientWithResponses, id uuid.UUID) (string, error) {

	resp, err := client.GetObjectPasswordId(context.Background(), id)
	if err != nil {
		return "", fmt.Errorf("error fetching %s item. Error: %v", sl.ProductName, err)
	}

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("error fetching %s item with status code: %d", sl.ProductName, resp.StatusCode)
	}

	if resp.ContentLength < 1 {
		return "", fmt.Errorf("password fetch unsuccessful")
	}

	defer resp.Body.Close()
	builder := new(strings.Builder)
	_, err = io.Copy(builder, resp.Body)
	if err != nil {
		return "", fmt.Errorf("error with reading %s response body: %v", sl.ProductName, err)
	} else {
		data := builder.String()

		var itemPassword ItemPassword
		err = json.Unmarshal([]byte(data), &itemPassword)
		if err != nil {
			fmt.Printf("error during Unmarshal(): %s", err)
		}

		if itemPassword.Success {
			return itemPassword.Data.Data, nil
		} else {
			return "", fmt.Errorf("password fetch unsuccessful")
		}
	}

}

func New(config string, custom interface{}) (salainen.SecretStorage, error) {

	storage := f{}
	err := storage.Init(custom)
	if err != nil {
		return nil, err
	}

	return &storage, nil
}

func Register(config string, custom interface{}) error {

	storage, err := New(config, custom)
	if err != nil {
		return err
	}
	salainen.AddSecretStorage(providerName, storage)

	return nil
}
