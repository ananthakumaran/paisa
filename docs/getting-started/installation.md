---
description: "How to install Paisa, an open source personal finance manager"
---


# Installation

Paisa is available in two formats: a **Desktop Application** and a **CLI**
(Command Line Interface). Both provide the same list of features, with
the primary difference being how the user interface is launched.

## Desktop Application

=== "Linux"

    * Download the prebuilt [binary](https://github.com/ananthakumaran/paisa/releases/latest) named `paisa-app-linux-amd64.deb`
    * You can install it either by double clicking the deb file or run the following commands in a Terminal

    ```console
    # cd ~/Downloads
    # sudo dpkg -i paisa-app-linux-amd64.deb
    ```

=== "Mac"

    * Download the prebuilt [binary](https://github.com/ananthakumaran/paisa/releases/latest) named `paisa-app-macos-amd64.dmg`
    * Open the dmg file and drag the Paisa app into Application folder
    * Since the app is not signed[^1], Mac will show a warning when
    you try to open the app. You can check the
    [support](https://support.apple.com/en-us/HT202491) page for more
    details. If you don't get any option to open the app, go to the
    Application folder, right click on the icon and select
    open. Usually, this should present you with an option to open.
    * Paisa will store all your journals, configuration files, and other
    related files in a folder named `paisa` which will be located in your
    `Documents` folder. When you open the app on your Mac for the
    first time, a permission dialog will appear. Click Allow, then close and reopen the app.

=== "Windows"

    * Download the prebuilt [binary](https://github.com/ananthakumaran/paisa/releases/latest) named `paisa-app-windows-amd64.exe`
    * Since the app is not signed[^1], Windows will show multiple
    warnings, You might have to click `Keep anyway`, `More info`, `Run
    anyway` etc.
    * Paisa will store all your journals, configuration files, and other
    related files in a folder named `paisa` which will be located in your
    `Documents` folder.

## Third Party Hosted Options [^2]

=== "PikaPods"

    <a href="https://www.pikapods.com/pods?run=paisa" target="_blank"
    rel="noopener" markdown>![Run on PikaPods](../images/pika-pods.svg)</a><br />
    PikaPods lets you run Paisa on a server with a few clicks. It handles app upgrades,
    backup and hosting for you. Make sure to setup a [user
    account](../reference/user-authentication.md) post installation.

## CLI

=== "Linux"

    * Download the prebuilt [binary](https://github.com/ananthakumaran/paisa/releases/latest) named `paisa-cli-linux-amd64`
    * Run the following commands in a Terminal

    ```console
    # cd ~/Downloads
    # mv paisa-cli-linux-amd64 paisa
    # chmod u+x paisa
    # mv paisa /usr/local/bin
    ```

=== "Mac"

    * Download the prebuilt [binary](https://github.com/ananthakumaran/paisa/releases/latest) named `paisa-cli-macos-amd64`
    * Run the following commands in a Terminal

    ```console
    # cd ~/Downloads
    # mv paisa-cli-macos-amd64 paisa
    # chmod u+x paisa
    # xattr -dr com.apple.quarantine paisa
    # mv paisa /usr/local/bin
    ```

=== "Windows"

    * Download the prebuilt [binary](https://github.com/ananthakumaran/paisa/releases/latest) named `paisa-cli-windows-amd64.exe`
    * Since the binary is not signed[^1] with a certificate, you might get
    a warning from Windows. You would have to click `keep anyway`.
    * Run the following command in a Windows PowerShell. Make sure
    you are in the correct folder. You should see something like `PS C:\Users\yourname>`

    ```pwsh-session
    PS C:\Users\john> mv .\Downloads\paisa-cli-windows-amd64.exe .\paisa.exe
    ```

    * The `paisa.exe` binary will be placed in the user's home directory. You
    can access it via PowerShell. Just open a new PowerShell window,
    which will start in the home directory. Then, you can run `.\paisa.exe serve`

!!! tip

    Paisa depends on the **[ledger](https://www.ledger-cli.org/download.html)** binary. The prebuilt
    paisa binaries come with an embedded ledger binary and will use it if
    it's not already installed on your system. If you prefer to install the
    ledger yourself, follow the installation instructions on [ledger](https://www.ledger-cli.org/download.html) site.

## CLI Quick Start

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

Go to [http://localhost:7500](http://localhost:7500). Read the [tutorial](./tutorial.md) to learn
more.

## Docker

Paisa CLI is available on [dockerhub](https://hub.docker.com/r/ananthakumaran/paisa). The default image only
supports ledger. `paisa:<version>-hledger`,
`paisa:<version>-beancount` or `paisa:<version>-all` image variants
can be used if you want to use paisa with others.

=== "Linux"

    ```console
    # mkdir -p /home/john/Documents/paisa/
    # docker run -p 7500:7500 -v /home/john/Documents/paisa/:/root/Documents/paisa/ ananthakumaran/paisa:latest
    ```

=== "Mac"

    ```console
    # mkdir -p /Users/john/Documents/paisa/
    # docker run -p 7500:7500 -v /Users/john/Documents/paisa/:/root/Documents/paisa/ ananthakumaran/paisa:latest
    ```

## Nix Flake

Paisa CLI is available as a nix flake.

=== "Linux"

    ```console
    # nix profile install github:ananthakumaran/paisa
    ```

=== "Mac"

    ```console
    # nix profile install github:ananthakumaran/paisa
    ```

[^1]: I offer Paisa as a free app, and I don't generate any revenue
      from it. Code signing would require me to pay $99 for Mac and
      approximately $300 for Windows each and every year to get the
      necessary certificates. I can't justify spending that much for
      an app that doesn't generate any income. Unfortunately, as a
      result, you would have to jump through hoops to get it working.

[^2]: As the name implies, these are third party hosting solutions
    operated by independent companies. I may receive affiliate
    compensation for linking to their websites.
