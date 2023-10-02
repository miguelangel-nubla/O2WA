# OAuth2-WebActions Server

Bridge the gap between command-line and web interfaces with OAuth2-WebActions (O2WA).

This tool transforms CLI workflows into web actions while leveraging the security of OAuth2.0 and OIDC.

## Features

1. **Your CLI on the Web**: Convert command-line tasks into web actions, making powerful commands accessible to non-tech-savvy users.
  
2. **Robust Security**: OAuth2.0 and OIDC integration ensure the secure execution of web-based commands.

3. **Dynamic Endpoints**: The server adapts web endpoints to commands via a JSON configuration, ensuring flexibility without altering the core code.

4. **Interactive Feedback**: Get HTML feedback for every command executed through the web, enhancing the user experience.

## Why Choose O2WA?

1. **Simplify for Users**: Make intricate tasks user-friendly by transitioning them from the CLI to the web.

2. **Unified Command Management**: Update commands centrally, maintaining consistency for all users.

3. **Safety First**: OAuth2 and OIDC-backed authentication ensure the safety of every command.

4. **Scalable**: Seamlessly incorporate new CLI tasks into the web interface via dynamic endpoints.

Consider O2WA if you're looking to expand your CLI operations to a wider audience securely and efficiently.

---

## Quickstart:

For those who like a straightforward setup without the need for containerization, here's a guide to quickly get a O2WA server operational:

### Step 1: Install `o2wa`

You have two options:

**Option A: From GitHub Releases**

1. Visit the [GitHub releases page](https://github.com/miguelangel-nubla/o2wa/releases).
2. Depending on your operating system and architecture, download the appropriate release asset. For instance:
   - For macOS with arm64: `O2WA_Darwin_arm64.tar.gz`
   - For Linux with x86_64: `O2WA_Linux_x86_64.tar.gz`
   ... and so on.
3. Extract the downloaded file.
4. Move the binary to a location in your system's `$PATH`, or you can execute it directly from the extracted folder.

**Option B: Using Go**

If you have Go installed, you can directly install the package:

```bash
go install github.com/miguelangel-nubla/o2wa@latest
```

Ensure your `$GOPATH/bin` directory is in your `$PATH` to run the installed binary from anywhere.

### Step 2: Set Up Configuration

In the directory where `o2wa` is installed or your working directory, you'll find a sample `config.json` configuration file. Tailor this file to your environment:

Ensure you update the OAuth2 endpoint URLs, client credentials, and define any custom commands as needed.

Use the provided /public endpoint that runs `echo Hello world!` as a starting point.

### Step 3: Launch `o2wa`

With everything configured, you can start `o2wa`:

```bash
o2wa
```

### Step 4: Access the Web Interface

Open a browser and navigate to the `o2wa` web interface:

```
http://localhost:8080  # Make sure to adjust the port if you've modified it in config.json
```

Then, follow the on-screen guidance to authenticate and execute your defined commands.

---

## Docker Deployment

Sample config.json is provided in the repository. Deploy with:

```bash
docker run \
    -p 8080:8080 \
    -v $PWD/config.json:/config.json \
    ghcr.io/miguelangel-nubla/o2wa:latest
```

### Initialization Script for Custom Tools

In scenarios where `o2wa`'s custom commands require special tools or setups, you can introduce an initialization script that runs before the main program.

1. Prepare your initialization script. For instance, name it `custom-init.sh`.
   
2. Launch the Docker container, mounting your script and setting `INIT_SCRIPT_PATH``:
    ```bash
    docker run \
        -p 8080:8080 \
        -v $PWD/config.json:/config.json \
        -v /path/to/your/custom-init.sh:/custom-init.sh \
        -e INIT_SCRIPT_PATH=/custom-init.sh \
        ghcr.io/miguelangel-nubla/o2wa
    ```

Your script will execute before `o2wa`, ensuring the necessary tools or configurations are in place.

### Custom CA Certificates

`o2wa` might interact with services using certificates from a private/internal CA. For smooth, secure interactions, trust these certificates.

For Docker deployments, mount your CA certificates at `/trusted-ca-certs`:

```bash
docker run \
    -p 8080:8080 \
    -v $PWD/config.json:/config.json \
    -v /path/to/your/ca/certs:/trusted-ca-certs \
    ghcr.io/miguelangel-nubla/o2wa
```

## Contributing

Enhance O2WA through documentation improvements, added features, or issue resolutions. Your contributions are valued!