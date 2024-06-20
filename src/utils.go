package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base32"
	"encoding/binary"
    "io"
	"bytes"
    "mime/multipart"
	"time"

    "github.com/AlecAivazis/survey/v2"
	"github.com/mgutz/ansi"
)

// Define a struct for the known parts of the JSON structure
type Plugins struct {
	PreComp     map[string][]string `json:"pre_comp"`
	Keying      map[string][]string `json:"keying"`
	Execution   map[string][]string `json:"execution"` 
	PayloadMods map[string][]string `json:"payload_mods"`
	PostComp    map[string][]string `json:"post_comp"`
}

// Struct to unmarshal JSON response
type StatusResponse struct {
	Status string `json:"status"`
}

type Log int64

const (
    logError Log = iota
    logInfo
    logStatus
    logInput
	logSuccess
	logSection
	logSubSection
)

// Function to print logs
func printLog(log Log, text string) {
	switch log {
	case logError:
		fmt.Printf("[%s] %s %s\n", ansi.ColorFunc("red")("!"), ansi.ColorFunc("red")("ERROR:"), ansi.ColorFunc("cyan")(text))
	case logInfo:
		fmt.Printf("[%s] %s\n", ansi.ColorFunc("blue")("i"), text)
	case logStatus:
		fmt.Printf("[*] %s\n", text)
	case logInput:
		fmt.Printf("[%s] %s", ansi.ColorFunc("yellow")("?"), text)
	case logSuccess:
		fmt.Printf("[%s] %s\n", ansi.ColorFunc("green")("+"), text)
	case logSection:
		fmt.Printf("\t[%s] %s\n", ansi.ColorFunc("yellow")("-"), text)
	case logSubSection:
		fmt.Printf("\t\t[%s] %s\n", ansi.ColorFunc("magenta")(">"), text)
	}
}

// Function to get plugins from the server
func getPlugins(ip string, port int) (Plugins, error) {
	var plugins Plugins

	printLog(logInfo, fmt.Sprintf("%s %s", ansi.ColorFunc("default+hb")("Server IP: "), ansi.ColorFunc("cyan")(ip)))
	printLog(logInfo, fmt.Sprintf("%s %s", ansi.ColorFunc("default+hb")("Server Port: "), ansi.ColorFunc("cyan")(fmt.Sprintf("%d", port))))

	url := fmt.Sprintf("http://%s:%d/api/v1/plugins", ip, port)
	resp, err := http.Get(url)
	if err != nil {
		return plugins, fmt.Errorf("failed to fetch plugins: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return plugins, fmt.Errorf("server returned non-200 status: %d %s", resp.StatusCode, resp.Status)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return plugins, fmt.Errorf("failed to read response body: %v", err)
	}

	if err := json.Unmarshal(body, &plugins); err != nil {
		return plugins, fmt.Errorf("failed to parse known parts of JSON: %v", err)
	}

	return plugins, nil
}

func printPlugins(label string, plugins map[string][]string) {
	printLog(logSuccess, fmt.Sprintf("%s:", ansi.ColorFunc("default+hb")(label)))

	for key, elements := range plugins {
		printLog(logSection, fmt.Sprintf("%s:", ansi.ColorFunc("default+hb")(key)))
		for _, element := range elements {
			printLog(logSubSection, fmt.Sprintf("%s", ansi.ColorFunc("cyan")(element)))
		}
	}
}

// Function to display plugins in a readable format
func displayPlugins(plugins Plugins) {
	// Display the known parts
	printLog(logInfo, "Known Plugins retrieved from server")
	printPlugins("Keying", plugins.Keying)
	printPlugins("Payload Mods", plugins.PayloadMods)
	printPlugins("Execution", plugins.Execution)
	printPlugins("Pre Compilation", plugins.PreComp)
	printPlugins("Post Compilation", plugins.PostComp)
}

// Retrieve config from file
func readConfig(config string) (map[string][]string, error) {
	file, err := os.Open(config)
	if err != nil {
		return nil, fmt.Errorf("failed to open config file: %v", err)
	}
	defer file.Close()

	var conf map[string][]string
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&conf); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %v", err)
	}

	return conf, nil
}

func Checkboxes(label string, opts []string, oneOption bool) []string {
    res := []string{}
    prompt := &survey.MultiSelect{
        Message: label,
        Options: opts,
    }
	if(oneOption) {
		survey.AskOne(prompt, &res, 
			survey.WithValidator(survey.Required), 
			survey.WithValidator(survey.MinItems(1)), 
			survey.WithValidator(survey.MaxItems(1)), 
			survey.WithRemoveSelectAll(), 
			survey.WithRemoveSelectNone(), 
			survey.WithIcons(func(icons *survey.IconSet) {
				// you can set any icons
				icons.Question.Text = fmt.Sprintf("[%s]", ansi.ColorFunc("yellow")("?"))
				// for more information on formatting the icons, see here: https://github.com/mgutz/ansi#style-format
				icons.Question.Format = ""
			}))
	} else {
    	survey.AskOne(prompt, &res, 
			survey.WithRemoveSelectAll(), 
			survey.WithRemoveSelectNone(), 
			survey.WithIcons(func(icons *survey.IconSet) {
				// you can set any icons
				icons.Question.Text = fmt.Sprintf("[%s]", ansi.ColorFunc("yellow")("?"))
				// for more information on formatting the icons, see here: https://github.com/mgutz/ansi#style-format
				icons.Question.Format = ""
			}))
	}
    return res
}

// Present menu and retrieve paths for plugins from sections
func getOptions(prefix string, plugins map[string][]string, label string, oneOption bool) []string {
	var options []string

	// Get options
	for key, elements := range plugins {
		for _, element := range elements {
			plugin := "/" + prefix + "/" + key + "/" + element
			options = append(options, plugin)
		}
	}

	return Checkboxes(label, options, oneOption)
}

// Ask user for configuration
func getConfig(ip string, port int) (map[string][]string, error) {
	userConfig := make(map[string][]string)

	plugins, err := getPlugins(ip, port)
	if err != nil {
		return userConfig, err
	}

	printLog(logInfo, ansi.ColorFunc("default+hb")("Plugins Selection: "))
	userConfig["keying"] 		= getOptions("keying", plugins.Keying, "Keying", false)
	userConfig["payload_mods"] 	= getOptions("payload_mods", plugins.PayloadMods, "Payload Mods", false)
	userConfig["execution"] 	= getOptions("execution", plugins.Execution, "Execution", true)
	userConfig["pre_comp"] 		= getOptions("pre_comp", plugins.PreComp, "Pre Compilation", false)
	userConfig["post_comp"] 	= getOptions("post_comp", plugins.PostComp, "Post Compilation", false)

	return userConfig, nil
}

// Save configuration file
func saveConfigFile(config map[string][]string) {
	printLog(logInput, ansi.ColorFunc("default+hb")("Save config to file [y/n]? "))
	
	fmt.Print("", ansi.ColorCode("cyan"))

	var saveFile rune
	_, err := fmt.Scanf("%c", &saveFile)
	if err != nil {
		printLog(logError, fmt.Sprintf("%v", err))
	}
	
	fmt.Print("", ansi.ColorCode("reset"))

	if (saveFile == 'y') {
		jsonData, err := json.MarshalIndent(config, "", "    ")
		if err != nil {
			printLog(logError, fmt.Sprintf("%v", err))
		}

		err = os.WriteFile("config.json", jsonData, 0644)
		if err != nil {
			printLog(logError, fmt.Sprintf("%v", err))
		}

		printLog(logSuccess, "Configuration file successfully written to config.json")
	}
}

// Generate UUID to be used for loader processing
func generateID() (string, error) {
	// Generate a random 16-bit integer
	var randomInt uint16
	err := binary.Read(rand.Reader, binary.LittleEndian, &randomInt)
	if err != nil {
		return "", fmt.Errorf("error generating random 16-bit integer: %w", err)
	}

	// Convert the random integer to a byte slice
	randomBytes := make([]byte, 2)
	binary.LittleEndian.PutUint16(randomBytes, randomInt)

	// Hash the byte slice using SHA-256
	hasher := sha256.New()
	_, err = hasher.Write(randomBytes)
	if err != nil {
		return "", fmt.Errorf("error hashing the random bytes: %w", err)
	}
	hash := hasher.Sum(nil)

	// Encode the hash in Base32
	base32Hash := base32.StdEncoding.EncodeToString(hash)

	return base32Hash, nil
}

func UploadMultipartFile(client *http.Client, uri, key, path string) (*http.Response, error) {
    body, writer := io.Pipe()

    req, err := http.NewRequest(http.MethodPost, uri, body)
    if err != nil {
        return nil, err
    }

    mwriter := multipart.NewWriter(writer)
    req.Header.Add("Content-Type", mwriter.FormDataContentType())

    errchan := make(chan error)

    go func() {
        defer close(errchan)
        defer writer.Close()
        defer mwriter.Close()

        w, err := mwriter.CreateFormFile(key, path)
        if err != nil {
            errchan <- err
            return
        }

        in, err := os.Open(path)
        if err != nil {
            errchan <- err
            return
        }
        defer in.Close()

        if written, err := io.Copy(w, in); err != nil {
            errchan <- fmt.Errorf("error copying %s (%d bytes written): %v", path, written, err)
            return
        }

        if err := mwriter.Close(); err != nil {
            errchan <- err
            return
        }
    }()

    resp, err := client.Do(req)
    merr := <-errchan

    if err != nil || merr != nil {
        return resp, fmt.Errorf("http error: %v, multipart error: %v", err, merr)
    }

    return resp, nil
}

func sendPayload(ip string, port int, id string, payloadFile string) {
	printLog(logInfo, fmt.Sprintf("%s %s", ansi.ColorFunc("default+hb")("Payload file: "), ansi.ColorFunc("cyan")(payloadFile)))
	printLog(logInfo, fmt.Sprintf("%s %s", ansi.ColorFunc("default+hb")("Client ID: "), ansi.ColorFunc("cyan")(id)))

	// Define the URI
	uri := fmt.Sprintf("http://%s:%d/api/v1/payload/upload/%s", ip, port, id)

	// Server expect key to be payload
	key := "payload"

	client := &http.Client{}

	// Upload the file to the server
	resp, err := UploadMultipartFile(client, uri, key, payloadFile)
	if err != nil {
		printLog(logError, fmt.Sprintf("Failed to upload payload: %v", err))
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		printLog(logError, fmt.Sprintf("Failed to upload payload, server responded with status: %s", resp.Status))
		return
	}

	printLog(logInfo, fmt.Sprintf("%s", ansi.ColorFunc("default+hb")("Payload uploaded successfully")))
}

// Send request to server to generate loader and retrieves the loader
func generateLoader(ip string, port int, id string, config map[string][]string) {
	// Define the URI
	uri := fmt.Sprintf("http://%s:%d/api/v1/payload/generate/%s", ip, port, id)
	
	// Convert config map to JSON
	jsonData, err := json.Marshal(config)
	if err != nil {
		printLog(logError, fmt.Sprintf("Failed to marshal config to JSON: %v", err))
		return
	}

	// Create a new POST request with the JSON payload
	req, err := http.NewRequest(http.MethodPost, uri, bytes.NewBuffer(jsonData))
	if err != nil {
		printLog(logError, fmt.Sprintf("Failed to create request: %v", err))
		return
	}

	// Set the content type to application/json
	req.Header.Set("Content-Type", "application/json")

	// Create a new HTTP client and send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		printLog(logError, fmt.Sprintf("Failed to send request: %v", err))
		return
	}
	defer resp.Body.Close()

	// Check the response status
	if resp.StatusCode != http.StatusOK {
		printLog(logError, fmt.Sprintf("Server responded with status: %s", resp.Status))
		return
	}

	printLog(logInfo, fmt.Sprintf("%s", ansi.ColorFunc("default+hb")("Loader generation requested successfully")))
}

// Requests loader from server every second
func requestLoader(ip string, port int, id string) {
	// Define the URIs
	statusUri := fmt.Sprintf("http://%s:%d/api/v1/payload/status/%s", ip, port, id)
	resultUri := fmt.Sprintf("http://%s:%d/api/v1/payload/result/%s", ip, port, id)

	// For every second make get request on status until response is finished
	for {
		// Make GET request on statusUri
		resp, err := http.Get(statusUri)
		if err != nil {
			printLog(logError, fmt.Sprintf("%v", err))
			time.Sleep(1 * time.Second)
			continue
		}

		defer resp.Body.Close()

		// Decode the JSON response
		var statusResponse StatusResponse
		if err := json.NewDecoder(resp.Body).Decode(&statusResponse); err != nil {
			printLog(logError, fmt.Sprintf("%v", err))
			time.Sleep(1 * time.Second)
			continue
		}

		// Check if the status is "Finished"
		if statusResponse.Status == "Finished" {
			printLog(logStatus, fmt.Sprintf("%s", ansi.ColorFunc("default+hb")("Status is Finished")))

			// Make GET request on resultUri to download the result
			resultResp, err := http.Get(resultUri)
			if err != nil {
				printLog(logError, fmt.Sprintf("%v", err))
				return
			}

			defer resultResp.Body.Close()

			// Create the result file
			out, err := os.Create("loader.exe")
			if err != nil {
				printLog(logError, fmt.Sprintf("%v", err))
				return
			}
			defer out.Close()

			// Write the response body to the file
			_, err = io.Copy(out, resultResp.Body)
			if err != nil {
				printLog(logError, fmt.Sprintf("%v", err))
				return
			}

			printLog(logSuccess, fmt.Sprintf("%s", ansi.ColorFunc("default+hb")("Result file downloaded successfully")))

			return
		}

		// Wait for 1 second before making the next request
		time.Sleep(1 * time.Second)
	}
}
