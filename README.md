<div align="center">Simplifying the deployment for smaller applications</div>

###

Burp is a deployment tool that is catered towards smaller applications that uses Docker. 
It enables developers to remotely deploy their application without having to connect to their 
server, spawning the required Docker containers all at once.

>  "Deploying small applications doesn't have to be crazy."

To make your application Burp-compatible, you have to create a `burp.toml` file on your project's 
root folder with the following specifications:
```toml
version = 1.0

[service]
name = "burp"
build = "."
repository = "github.com/ShindouMihou/burp" # Burp will clone and pull the source code from here
```

And the application should now be Burp-compatiable. You can then deploy the application by running the 
command:
```shell
burp deploy --server [server_name]
```

To kill  all the containers that Burp spawned for the project, all you have to do is run the 
following command:
```shell
burp stop --server [server_name]
```

#### Dependencies

Burp supports spawning the required services that your application needs, such as Mongo, or Redis. To do so, 
you have to declare the dependencies that your application needs, which includes the image, and so  forth.

```toml
[[dependencies]]
name = "redis"
image = "redis"
ports = [[6379,6379]]
```

##### Functional Burp

Burp also includes multiple functions such as `Random` and `Use` which allows us to generate values such as long,  
randomized strings that can be used for passwords. For example, if we want to secure our Redis, we can use the following:
```toml
[[dependencies]]
name = "redis"
image = "redis"
command = "redis-server  --appendonly  yes --requirepass \"[burp: Random(12) AS redis_pass]\""
ports = [[6379,6379]]
```

You can call functions by using the Burper syntax: `[burp: Function(args)]` and you can save the result of that function 
by including an `AS name` after the function call such as `[burp: Function(args) AS var]`.

To reuse the value further down the line, you can simply use the `Use` function:
```toml
[[dependencies.environment]]
REDIS_PASSWORD="[burp: Use(redis_pass)]"
```

##### Environment Replacements

Burp also supports replacing configuration properties from your `.env` file before deployment. To do so, you 
have to declare the `env` file that you are using, then declare the replacements that you want.

For example, if we have an `.env` file such as:
```dotenv
APPLICATION_NAME=burp
REDIS_URI=
```
```toml
[env]
file = "[burp].env"
output = ".env"

[[env.replacements]]
REDIS_URI="redis://root:[burp: Use(redis_pass)]@172.17.0.1:6379"
```

> As Burp overrides the file, we highly recommend setting the `file` into something other than `.env` such as 
> `[burp].env` which serves as a template.

##### Setting Up Burp

To set up Burp, you need to have Burp Agent installed on your server, and the easiest way to do so is to use Burp itself, 
you can do this by clone the Burp repository and downloading the [Burp CLI](https://github.com/ShindouMihou/burp/releases) onto 
your server, and following the steps:

1. Creating a folder named `/data/burp` somewhere, Burp will copy all uploaded files that you'll include from deployment over to that folder.
2. Rename the `.env.example` into a `burp.env` then configuring the file.
3Running the following command: `burp deploy --here`

There are two properties that Burp needs, and those are:
- `BURP_SECRET`: Akin to a password, this is needed to authenticate people to the agent. You have to hash this with argon2id since
the server will only need the hash.
- `BURP_SIGNATURE`: Akin to a username, this is used as an initial check over whether the request should be hash-checked. This is 
to reduce resources as requests that do not contain this signature are ignored.

You also need to create a `git.toml` and `registries.toml` if you want to access private repositories or pull images 
from the Dockerhub more than 100 times a day, please create them under the `data/` folder in the repository since 
Burp will copy them over to the server.

An example of a `git.toml` would be:
```toml
[[git]]
domain = "github.com"
username = "abc" # GitHub ignores this
password = "[access_token]"
```

An example of a `registries.toml` would be:
```toml
[[registry]]
domain = ""
username = ""
password = "" # Personal Token if DockerHub or GitHub
```

If you want to hash a password with argon2id, you can use the command from burp:
```shell
burp hash [password]
```

Once you have the agent on the server, you can register the server on your client by running the following command on your 
development environment (e.g. PC):
```shell
burp login
```

And it should ask you for the server address, the server name and the secret token plus signature. Make sure to give the 
non-hashed secret token, Burp will handle the rest afterward.

#### State of Burp
- [x] Burper
  - [x] Parsing of Burp Functions
  - [x] Adding Functions
  - [x] Processing of Functions
  - [x] Basic Functions (mathematics, hash, etc)
- [x] Burp.TOML
  - [x] Structures
  - [x] Reading of TOML file
- [x] Git Support
  - [x] Cloning of Git repository
- [x] Docker Support
  - [x] Pulling images
  - [x] Building images
  - [x] Spawning services
  - [x] Killing services
- [x] Agent Client
  - [x] Authentication
  - [x] Deploying Applications
  - [x] Stopping Applications
  - [x] Removing Applications
- [ ] CLI
  - [ ] `burp hash`
  - [ ] `burp login`
  - [ ] `burp deploy`
  - [ ] `burp stop`