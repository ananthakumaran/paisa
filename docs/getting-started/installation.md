# Installation

=== "Linux"

    * Download the prebuilt [binary](https://github.com/ananthakumaran/paisa/releases/latest) named `paisa-linux-amd64`
    * Run the following commands in a Terminal

    ```console
    # cd ~/Downloads
    # mv paisa-linux-amd64 paisa
    # chmod u+x paisa
    # mv paisa /usr/local/bin
    ```

=== "Mac"

    * Download the prebuilt [binary](https://github.com/ananthakumaran/paisa/releases/latest) named `paisa-macos-amd64`
    * Run the following commands in a Terminal

    ```console
    # cd ~/Downloads
    # mv paisa-macos-amd64 paisa
    # chmod u+x paisa
    # xattr -dr com.apple.quarantine paisa
    # mv paisa /usr/local/bin
    ```

=== "Windows"

    * Download the prebuilt [binary](https://github.com/ananthakumaran/paisa/releases/latest) named `paisa-windows-amd64.exe`
    * Since the binary is not signed with a certificate, you might get
    a warning from Windows. You would have to click `keep anyway`.
    * Run the following command in a Windows PowerShell. Make sure
    you are in the correct folder. You should see something like `PS C:\Users\yourname>`

    ```pwsh-session
    PS C:\Users\john> mv .\Downloads\paisa-windows-amd64.exe .\paisa.exe
    ```

    * The `paisa.exe` binary will be placed in the user's home directory. You
    can access it via PowerShell. Just open a new PowerShell window,
    which will start in the home directory. Then, you can run `.\paisa.exe serve`

!!! tip

    Paisa depends on the **[ledger](https://www.ledger-cli.org/download.html)** binary. The prebuilt
    paisa binaries come with an embedded ledger binary and will use it if
    it's not already installed on your system. If you prefer to install the
    ledger yourself, follow the installation instructions on [ledger](https://www.ledger-cli.org/download.html) site.

## Quick Start

Paisa will store all your journals, configuration files, and other
related files in a folder named `paisa` which will be located in your
`Documents` folder.

=== "Linux"

    ```console
    # paisa serve
    ```

=== "Mac"

    ```console
    # paisa serve
    ```

=== "Windows"

    ```pwsh-session
    PS C:\Users\john> .\paisa.exe serve
    ```

Go to [http://localhost:7500](http://localhost:7500). Read the [tutorial](./tutorial.md) to learn more.
