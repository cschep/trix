package trix

// getClient uses a Context and Config to retrieve a Token
import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"

	"google.golang.org/api/sheets/v4"

	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

//Trix holds the spreadsheetID and the sheetsService for a trix
type Trix struct {
	spreadsheetID string
	sheetsService *sheets.Service
}

//NewTrix returns a *Trix
func NewTrix(spreadsheetID string) (*Trix, error) {
	ctx := context.Background()

	b, err := ioutil.ReadFile("client_secret.json")
	if err != nil {
		return nil, fmt.Errorf("Unable to read client secret file: %v", err)
	}

	// If modifying these scopes, delete your previously saved credentials
	// at ./.credentials/sheets.googleapis.com-go-quickstart.json
	config, err := google.ConfigFromJSON(b, "https://www.googleapis.com/auth/spreadsheets")
	if err != nil {
		return nil, fmt.Errorf("Unable to parse client secret file to config: %v", err)
	}
	client := getClient(ctx, config)

	srv, err := sheets.New(client)
	if err != nil {
		return nil, fmt.Errorf("Unable to retrieve Sheets Client %v", err)
	}

	t := &Trix{spreadsheetID: spreadsheetID, sheetsService: srv}

	return t, nil
}

// then generate a Client. It returns the generated Client.
func getClient(ctx context.Context, config *oauth2.Config) *http.Client {
	cacheFile, err := tokenCacheFile()
	if err != nil {
		log.Fatalf("Unable to get path to cached credential file. %v", err)
	}
	tok, err := tokenFromFile(cacheFile)
	if err != nil {
		tok = getTokenFromWeb(config)
		saveToken(cacheFile, tok)
	}
	return config.Client(ctx, tok)
}

// getTokenFromWeb uses Config to request a Token.
// It returns the retrieved Token.
func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the "+
		"authorization code: \n%v\n", authURL)

	var code string
	if _, err := fmt.Scan(&code); err != nil {
		log.Fatalf("Unable to read authorization code %v", err)
	}

	tok, err := config.Exchange(oauth2.NoContext, code)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web %v", err)
	}

	return tok
}

// tokenCacheFile generates credential file path/filename.
// It returns the generated credential path/filename.
func tokenCacheFile() (string, error) {
	tokenCacheDir := filepath.Join("./", ".credentials")
	os.MkdirAll(tokenCacheDir, 0700)
	return filepath.Join(tokenCacheDir,
		url.QueryEscape("sheets.googleapis.com-go-quickstart.json")), nil
}

// tokenFromFile retrieves a Token from a given file path.
// It returns the retrieved Token and any read error encountered.
func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	t := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(t)
	defer f.Close()
	return t, err
}

// saveToken uses a file path to create a file and store the
// token in it.
func saveToken(file string, token *oauth2.Token) {
	fmt.Printf("Saving credential file to: %s\n", file)
	f, err := os.Create(file)
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}

//Get returns data from the sheet in the given range
func (t *Trix) Get(readRange string) (*sheets.ValueRange, error) {
	resp, err := t.sheetsService.Spreadsheets.Values.Get(t.spreadsheetID, readRange).Do()
	if err != nil {
		return nil, fmt.Errorf("Unable to retrieve data from sheet. %v", err)
	}

	return resp, nil
}

//Update updates data on the sheet in the given range
func (t *Trix) Update(updateRange string, values [][]interface{}) (*sheets.UpdateValuesResponse, error) {

	fmt.Println("trying to update with these values:", values)

	valuerange := &sheets.ValueRange{Values: values}
	resp, err := t.sheetsService.Spreadsheets.Values.Update(t.spreadsheetID, updateRange, valuerange).ValueInputOption("USER_ENTERED").Do()
	if err != nil {
		return nil, fmt.Errorf("Unable to UPDATE data ON sheet. %v", err)
	}

	return resp, nil
}

//InsertRow inserts a row of values at the bottom of the spreadsheet
func (t *Trix) InsertRow(values [][]interface{}) (*sheets.UpdateValuesResponse, error) {
	readRange := "RSVP!A:C"

	readResp, err := t.Get(readRange)
	if err != nil || len(readResp.Values) < 1 {
		log.Println("No Values.", err)
		return nil, err
	}

	for _, row := range readResp.Values {
		log.Println(row)
	}

	writeRow := len(readResp.Values) + 1
	updateRange := fmt.Sprintf("RSVP!A%d:C%d", writeRow, writeRow)

	updateResp, err := t.Update(updateRange, values)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	return updateResp, nil
}
