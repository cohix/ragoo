# 🍝 Ragoo /ræɡˈuː/
> Like the pasta dish.

## Overview
Create Agents and RAG workflows with little to no code.

- Experiment with Agents and RAG workflows with only YAML configuration. Iterate quickly.
- Easily import data from various sources, generate embeddings, store the data in a vector db. Keep that data continuously up to date.
- Extend your experiments into full applications by adding custom tools and (optionally) writing some code.

## Project Goals
- Create composable workflows which can be exposed as API endpoints or run on schedules.
- Achieve significant progress on Agent/RAG prototyping without needing to invest large amounts of time.
- Provide plugins for various services such as vector databases, LLMs, and data importers.
- Move past prototyping with code and plugin escape hatches. Assist in evaluating quality.
- Deploy "for real" and successfully serve production traffic. Eventually.

## Features
Implemented:
- YAML config for expressing Agent/RAG workflows, routes, and plugins
- Plugins for:
	- Importers: files
	- Vector DBs: DuckDB
	- LLM Services: Ollama
	- Embedders: Ollama
- HTTP server to expose workflows

Planned:
- Pluggable tools with LLM tool_choice support:
	- STDIN/STDOUT binaries
	- HTTP endpoints
	- Go packages
	- Workflows (LLM from one workflow can call another workflow)
- Observability (OpenTelemetry)
- Support for more types of plugins
- More extensive prompt templating support

## More information
See [the example config file](./ragoo.yaml) to get started. It uses docs from [Kubernetes the Hard Way](https://github.com/kelseyhightower/kubernetes-the-hard-way) (cloned locally) as example data.

Requires Go 1.22 to build. Default configuration uses Ollama for easy demonstration.

Early experimental phase.

Apache 2.0 Licensed.
Copyright Connor Hicks and contributors, 2024.