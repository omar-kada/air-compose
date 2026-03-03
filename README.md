# AirCompose

AirCompose is a lightweight gitops automation tool that handles **Docker compose** stacks deployment.  
Its purpose is to make deploying and updating a self-hosted environment **simple**, **fast**, and **reproducible**. without the need to have heavy tooling or dependencies.

## Summary

- [Requirements](#requirements)
- [Features](#features)
- [Getting started](#getting-started)

## Requirements

a Linux environement with `docker` installed

## Features

1. Auto Load and deploy services from a Git repository
2. No special syntax, compose stacks configuration are kept in their original form and are accessible to the user
3. Modern UI to manage the tool remotely
4. Get notifications about updates and changes
5. Easy service customization through the UI or a simple configuration file
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

3. **Create a `config.yaml` file** inside the specified `AIR_COMPOSE_DATA_PATH` and define the services you want to deploy (here a simple example) :

```yaml
ENV_VAR: value # will be available in all services
repo: 'https://github.com/omar-kada/air-compose-config'
cron: '*/10 * * * *'

services:
  service1:
    ENV_VAR: override value # will override global value for this service
    SERVICE_SPECIFIC_VAR: another_value

  service2:
    disabled: true # if disabled, service will not be deployed
```

4. **Run the stack** using :

```bash
docker compose up -d
```

Once the container starts, it will :

1. **Pull the stacks from the repo**
2. **Deploy or remove services** based on the configuration
3. **Schedule the next runs** based on `CRON_PERIOD`

When the stacks are updated in the repo, AirCompose will **redploy only the changed stacks** in the next scheduled run
