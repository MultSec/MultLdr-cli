package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

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

// Send request to server to generate loader and retrieves the loader