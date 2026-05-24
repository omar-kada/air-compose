# <img src="./frontend/public/air-compose.svg" width="32" style="vertical-align: middle;" /> AirCompose

AirCompose is a lightweight gitops automation tool that handles **Docker compose** stacks deployment.  
Its purpose is to make deploying and updating a self-hosted environment **simple**, **fast**, and **reproducible**. without the need to have heavy tooling or dependencies.

## Summary

- [Requirements](#requirements)
- [Features](#features)
- [Getting started](#getting-started)

## Requirements

a Linux environement with `docker` installed

## Features

1. Auto sync&deploy services from a Git repository
2. No special syntax, compose stacks configuration are kept in their original form and are accessible to the user
3. Modern UI to manage the tool remotely
4. Notifications on deployments and health (through [Shoutrrr](https://containrrr.dev/shoutrrr))
5. Override stacks' environement variables (for keys and token that should be kept local)
6. Works on any Docker-capable system

## Getting started

1. **Create a configuration repo** containing all your compose stacks with this structure (for example : [AirCompose Config](https://github.com/omar-kada/air-compose-config))

```
services/
├── service1/
|   ├── compose.yaml
|   └── .env
└── service2/
    ├── compose.yaml
    └── .env
```

2. **Copy the `compose.yaml` file your system** and fill the needed variables (make sure to read the comments about each variable), here are the main ones :

```yaml
AIR_COMPOSE_SERVICES_DIR: where the stack configuration will be stored
AIR_COMPOSE_DATA_PATH: path to the data directory, AirCompose will store config.yaml and DB files in this directory
```

3. **Run the stack** using :

```bash
docker compose up -d
```

4. **Register** a user and do the **initial configuration** through the UI (default port 5005)

the configuration is stored in a file named `config.yaml`, it can be changed either through the UI or by changing the file directly (consider restarting the container in this case).

5. Go to the configuration page to add custom environement variables for each service (optional)

6. Click on **Sync** to launch the first deployment (if not configured to run automatically)

The Sync will pull the stacks from the repo and deploy them. When the stacks are updated in the repo, AirCompose will **redploy only the changed stacks** in the next scheduled run (or if run manually)
