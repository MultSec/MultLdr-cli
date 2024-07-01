<div align="center">
  <img width="125px" src="assets/MultLdr-cli.png" />
  <h1>MultLdr-cli</h1>
  <br/>

  <p><i>MultLdr-cli is a command line interface for the MultLdr project, created by <a href="https://infosec.exchange/@Pengrey">@Pengrey</a>.</i></p>
  <br />
  
</div>

## Installation

* Run `go build` in the src directory of the project to build the binary.

* Or use the pre-built binary in the bin directory.

```bash
$ cd src
# Build for linux
$ GOOS=linux GOARCH=amd64 go build -o ../bin/multldr-cli_linux

# Build for windows (64-bit)
$ GOOS=windows GOARCH=amd64 go build -o ../bin/multldr-cli_64.exe

# Build for windows (32-bit)
$ GOOS=windows GOARCH=386 go build -o ../bin/multldr-cli_32.exe

# Build for mac
$ GOOS=darwin GOARCH=amd64 go build -o ../bin/multldr-cli_mac
```

## Demo

https://github.com/MultSec/MultLdr-cli/assets/55480558/7093f878-9f53-4ef2-abc9-cf2fd2391ada

## Documentation
For more information on how to use the MultLdr-cli, check the [documentation](https://multsec.github.io/docs/multldr-cli/)

## License
This project is licensed under the GNU General Public License - see the [LICENSE](LICENSE) file for details.