# Burp

*Deploying smaller applications should be less complicated.*

Burp is a deployment tool designed to simplify the deployment process for smaller applications that use Docker. It enables developers to remotely deploy their small applications without the need to connect directly to their server. Burp also takes care of handling the spawning of third-party services (containers) required for the application, making deployment hassle-free.

## Demo
![burp-deploy](https://github.com/ShindouMihou/burp/assets/69381903/4740b2a8-1720-434e-b624-fa55f307e6f4)

<details>
  <summary>Deploying with Pull</summary>
  
  ![burp-deploy-with-pull-resized](https://github.com/ShindouMihou/burp/assets/69381903/ba8c7d38-1035-42a6-8a2c-510fde9390d7)
</details>

## Getting Started
To get started with Burp, you have to install the command-line tool on both your server and development environment.
- [`GitHub Releases`](https://github.com/ShindouMihou/burp/releases)

Once you have the command-line tool installed, please refer to the following:
- [**Installing the Burp Agent on your remote server (TODO)**]()
- [**Setting up Burp CLI on your development machine (TODO)**]()

### Prerequisities
- Docker installed on both your development machine and your remote server.

## Deploying with Burp

To deploy with Burp, you have to make a [`burp.toml`](burp.toml). To know how to create one, you can refer to the 
following resources:
- [`GitHub Wiki (TODO)`](https://github.com/ShindouMihou/burp/wiki)

## License

Burp is distributed under the Apache 2.0 license. See [**LICENSE**](LICENSE) for more information.
