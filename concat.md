# Repository Structure

- **.github/**
  - **workflows/**
    - dev.yml
    - prd.yml
- **cmd/**
  - **laliga-matchfantasy-rabbitmq-publisher/**
    - main.go
- **config/**
  - config.go
  - db.go
  - listener.go
  - rmq.go
- **deploy/**
  - **eks/**
    - deployment_prod.yaml
    - deployment_qa.yaml
  - **gce/**
    - cloudbuild_production.yaml
    - cloudbuild_staging.yaml
  - **k8s/**
    - deployment.yaml
- **images/**
  - **laliga-matchfantasy-rabbitmq-publisher/**
    - deployment_dev.yaml
    - deployment_prd.yaml
- **publisher/**
  - connector.go
  - event.go
  - model.go
  - publisher.go
- README.md

---

# README.md

# laliga-matchfantasy-rabbitmq-publisher

---

## File: `.github/workflows/dev.yml`
- **File Size:** 2729 bytes
- **Last Modified:** Tue Apr 22 2025 10:28:27 GMT-0500 (Peru Standard Time)
- **Last Commit:** add prd yaml file to main

```
name: "[dev] K8S Rabbitmq Deploy"
defaults:
  run:
    shell: bash

env:
  DIGITALOCEAN_ACCESS_TOKEN: ${{ secrets.DIGITALOCEAN_ACCESS_TOKEN }}
  ENV: dev
  NAMESPACE: fanclash-dev
  REPO_NAME: laliga-matchfantasy-rabbitmq-publisher
on:
  push:
    branches:
      - main
  workflow_dispatch:

jobs:
  fanclash-rabbitmq-publisher:  # Create infrastructure for services on push to Main branch
    name: laliga-matchfantasy-rabbitmq-publisher
    runs-on: ubuntu-latest
    timeout-minutes: 15
    steps:
      - name: Checkout Code
        uses: actions/checkout@v2
        with:
          submodules: true
      - name: Install doctl 
        uses: digitalocean/action-doctl@v2
        with:
            token: ${{ secrets.DIGITALOCEAN_ACCESS_TOKEN }}
      - name: Log in to DO Container Registry 
        run: doctl registry login --expiry-seconds 600

      - name: Configure Kubectl for DOKS
        run: doctl kubernetes cluster kubeconfig save dev-fanclash # Replace <your-cluster-name> with your cluster's name

      - name: Build and Push Docker Image
        run: |
          REPO_NAME="${GITHUB_REPOSITORY##*/}"
          SHORT_SHA=$(echo $GITHUB_SHA | cut -c1-7)
          DOCKER_IMAGE="${REPO_NAME}:$SHORT_SHA"
          docker build -t $DOCKER_IMAGE .
          docker tag $DOCKER_IMAGE registry.digitalocean.com/gameon-ams3/$DOCKER_IMAGE
          docker push registry.digitalocean.com/gameon-ams3/$DOCKER_IMAGE

      - name: Update Image Tag in K8S Deployment
        run: |
          SHORT_SHA=$(echo $GITHUB_SHA | cut -c1-7)
          sed -i 's/TAG_PLACEHOLDER/'"$SHORT_SHA"'/g' $GITHUB_WORKSPACE/images/${{ env.REPO_NAME }}/deployment_${{ env.ENV }}.yaml

      - name: K8S Deploy - Deployment
        run: kubectl apply -f images/${{ env.REPO_NAME }}/deployment_${{ env.ENV }}.yaml

      - name: Check Deployment Health
        if: success()
        run: kubectl rollout status deployment/${{ env.REPO_NAME }} -n $NAMESPACE
        timeout-minutes: 3

      - name: Rollback Deployment
        if: failure()
        run: kubectl rollout undo deployment/${{ env.REPO_NAME }} -n $NAMESPACE
        timeout-minutes: 3

      - name: Slack Notification
        if: always()
        uses: rtCamp/action-slack-notify@v2
        env:
          SLACK_CHANNEL: staging-deployments
          SLACK_WEBHOOK: ${{ secrets.SLACK_WEBHOOK_STAGING_URL }}
          SLACK_ICON_EMOJI: ':gameon:'
          SLACK_USERNAME: GitHubAction
          SLACK_COLOR: ${{ job.status }} # Sets the color of the Slack notification bar to red for failures
          SLACK_TITLE: 'Staging Laliga Rabbitmq Publisher K8s deployment. Commit message: ${{ github.event.head_commit.message }}'
          SLACK_FOOTER: Powered By GameOn DevOps team
```

## File: `.github/workflows/prd.yml`
- **File Size:** 3000 bytes
- **Last Modified:** Tue Apr 22 2025 10:28:27 GMT-0500 (Peru Standard Time)
- **Last Commit:** add prd yaml file to main

```
name: "[prd] K8S Rabbitmq Deploy"
defaults:
  run:
    shell: bash

env:
  DIGITALOCEAN_ACCESS_TOKEN: ${{ secrets.DIGITALOCEAN_ACCESS_TOKEN }}
  ENV: prd
  NAMESPACE: prd-fanclash
  REPO_NAME: laliga-matchfantasy-rabbitmq-publisher
on:
  push:
    branches:
      - prd
  workflow_dispatch:

jobs:
  fanclash-rabbitmq-publisher:  # Create infrastructure for services on push to Main branch
    name: laliga-matchfantasy-rabbitmq-publisher
    runs-on: ubuntu-latest
    timeout-minutes: 15
    steps:
      - name: Checkout Code
        uses: actions/checkout@v2
        with:
          submodules: true
      - name: Install doctl 
        uses: digitalocean/action-doctl@v2
        with:
            token: ${{ secrets.DIGITALOCEAN_ACCESS_TOKEN }}
      - name: Log in to DO Container Registry 
        run: doctl registry login --expiry-seconds 600

      - name: Configure Kubectl for DOKS
        run: doctl kubernetes cluster kubeconfig save prd-fanclash # Replace <your-cluster-name> with your cluster's name

      - name: Build and Push Docker Image
        run: |
          REPO_NAME="${GITHUB_REPOSITORY##*/}"
          SHORT_SHA=$(echo $GITHUB_SHA | cut -c1-7)
          DOCKER_IMAGE="${REPO_NAME}:$SHORT_SHA"
          docker build -t $DOCKER_IMAGE .
          # Tagging
          docker tag $DOCKER_IMAGE registry.digitalocean.com/gameon-ams3/$DOCKER_IMAGE
          docker tag $DOCKER_IMAGE registry.digitalocean.com/gameon-ams3/laliga-matchfantasy-rabbitmq-publisher:prd
          # Pushing
          docker push registry.digitalocean.com/gameon-ams3/$DOCKER_IMAGE
          docker push registry.digitalocean.com/gameon-ams3/laliga-matchfantasy-rabbitmq-publisher:prd

      - name: Update Image Tag in K8S Deployment
        run: |
          SHORT_SHA=$(echo $GITHUB_SHA | cut -c1-7)
          sed -i 's/TAG_PLACEHOLDER/'"$SHORT_SHA"'/g' $GITHUB_WORKSPACE/images/laliga-matchfantasy-rabbitmq-publisher/deployment_${{ env.ENV }}.yaml

      - name: K8S Deploy - Deployment
        run: kubectl apply -f images/${{ env.REPO_NAME }}/deployment_${{ env.ENV }}.yaml

      - name: Check Deployment Health
        if: success()
        run: kubectl rollout status deployment/${{ env.REPO_NAME }} -n $NAMESPACE
        timeout-minutes: 3

      - name: Rollback Deployment
        if: failure()
        run: kubectl rollout undo deployment/${{ env.REPO_NAME }} -n $NAMESPACE
        timeout-minutes: 3

      - name: Slack Notification
        if: always()
        uses: rtCamp/action-slack-notify@v2
        env:
          SLACK_CHANNEL: production-deployments
          SLACK_WEBHOOK: ${{ secrets.SLACK_WEBHOOK_PRD_URL }}
          SLACK_ICON_EMOJI: ':gameon:'
          SLACK_USERNAME: GitHubAction
          SLACK_COLOR: ${{ job.status }} # Sets the color of the Slack notification bar to red for failures
          SLACK_TITLE: 'Prd Laliga Rabbitmq Publisher K8s deployment. Commit message: ${{ github.event.head_commit.message }}'
          SLACK_FOOTER: Powered By GameOn DevOps team
```

## File: `README.md`
- **File Size:** 41 bytes
- **Last Modified:** Tue Apr 22 2025 10:28:27 GMT-0500 (Peru Standard Time)
- **Last Commit:** first commit

```
# laliga-matchfantasy-rabbitmq-publisher

```

## File: `cmd/laliga-matchfantasy-rabbitmq-publisher/main.go`
- **File Size:** 1581 bytes
- **Last Modified:** Tue Apr 22 2025 10:28:27 GMT-0500 (Peru Standard Time)
- **Last Commit:** update the code for the new name

```
package main

import (
	"database/sql"
	"fmt"

	"github.com/gameon-app-inc/laliga-matchfantasy-rabbitmq-publisher/config"
	"github.com/gameon-app-inc/laliga-matchfantasy-rabbitmq-publisher/publisher"
	_ "github.com/jackc/pgx/stdlib"
	"github.com/lib/pq"
	"github.com/sirupsen/logrus"
)

func openDB() (*sql.DB, error) {
	// Start a database connection.
	db, err := sql.Open("pgx", config.DatabaseURL())
	if err != nil {
		return nil, err
	}

	// Actually test the connection against the database, so we catch
	// problematic connections early.
	if err := db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}

func openListener() (*pq.Listener, error) {
	listener := pq.NewListener(
		config.DatabaseURL(),
		config.ListenerMinReconnectInterval(),
		config.ListenerMaxReconnectInterval(),
		func(ev pq.ListenerEventType, err error) {
			if err != nil {
				logrus.WithError(err).Error("listener error occured")
			}
		})

	if err := listener.Listen(config.ListenerChannelName()); err != nil {
		logrus.WithField("channel", config.ListenerChannelName()).Error("failed to listen channel")
		return nil, err
	}

	return listener, nil
}

func main() {
	db, err := openDB()
	if err != nil {
		panic(fmt.Sprintf("failed to connect database: %v", err))
	}
	defer db.Close()

	listener, err := openListener()
	if err != nil {
		panic(fmt.Sprintf("failed to open listener: %v", err))
	}
	defer listener.Close()

	// run publishing
	done := make(chan bool)
	pub := publisher.NewEventPublisher(config.RMQConnectionURL(), config.RMQExchanges(), db, listener)
	pub.Start()

	<-done
}

```

## File: `config/config.go`
- **File Size:** 1437 bytes
- **Last Modified:** Tue Apr 22 2025 10:28:27 GMT-0500 (Peru Standard Time)
- **Last Commit:** Refactor code for improved performance and readability

```
package config

import (
	"time"

	"github.com/caarlos0/env"
)

// Represents a structure with all env variables needed by the backend.
var cfg struct {
	DatabaseUser                 string        `env:"DATABASE_USER,required"`
	DatabasePassword             string        `env:"DATABASE_PASSWORD,required"`
	DatabaseHost                 string        `env:"DATABASE_HOST,required"`
	DatabasePort                 int           `env:"DATABASE_PORT" envDefault:"5432"`
	DatabaseName                 string        `env:"DATABASE_NAME,required"`
	DatabaseSSLMode              string        `env:"DATABASE_SSLMODE" envDefault:"disable"`
	RMQHost                      string        `env:"RMQ_HOST,required"`
	RMQPort                      int           `env:"RMQ_PORT,required"`
	RMQVHost                     string        `env:"RMQ_VHOST,required"`
	RMQUser                      string        `env:"RMQ_USER,required"`
	RMQPassword                  string        `env:"RMQ_PASSWORD,required"`
	RMQExchanges                 []string      `env:"RMQ_EXCHANGES,required"`
	ListenerMinReconnectInterval time.Duration `env:"LISTENER_MIN_RECONNECT_INTERVAL" envDefault:"5s"`
	ListenerMaxReconnectInterval time.Duration `env:"LISTENER_MAX_RECONNECT_INTERVAL" envDefault:"30s"`
	ListenerChannelName          string        `env:"LISTENER_CHANNEL_NAME" envDefault:"amqp_events"`
}

func init() {
	if err := env.Parse(&cfg); err != nil {
		panic(err)
	}
}

```

## File: `config/db.go`
- **File Size:** 270 bytes
- **Last Modified:** Tue Apr 22 2025 10:28:27 GMT-0500 (Peru Standard Time)
- **Last Commit:** fix for sslmode

```
package config

import "fmt"

func DatabaseURL() string {
	return fmt.Sprintf("host=%s port=%d dbname=%s user=%s password=%s sslmode=%s",
		cfg.DatabaseHost,
		cfg.DatabasePort,
		cfg.DatabaseName,
		cfg.DatabaseUser,
		cfg.DatabasePassword,
		cfg.DatabaseSSLMode,
	)
}

```

## File: `config/listener.go`
- **File Size:** 298 bytes
- **Last Modified:** Tue Apr 22 2025 10:28:27 GMT-0500 (Peru Standard Time)
- **Last Commit:** Refactor code for improved performance and readability

```
package config

import (
	"time"
)

func ListenerMinReconnectInterval() time.Duration {
	return cfg.ListenerMinReconnectInterval
}

func ListenerMaxReconnectInterval() time.Duration {
	return cfg.ListenerMaxReconnectInterval
}

func ListenerChannelName() string {
	return cfg.ListenerChannelName
}

```

## File: `config/rmq.go`
- **File Size:** 252 bytes
- **Last Modified:** Tue Apr 22 2025 10:28:27 GMT-0500 (Peru Standard Time)
- **Last Commit:** Refactor code for improved performance and readability

```
package config

import "fmt"

func RMQConnectionURL() string {
	return fmt.Sprintf("amqp://%s:%s@%s:%d/%s",
		cfg.RMQUser,
		cfg.RMQPassword,
		cfg.RMQHost,
		cfg.RMQPort,
		cfg.RMQVHost,
	)
}

func RMQExchanges() []string {
	return cfg.RMQExchanges
}

```

## File: `deploy/eks/deployment_prod.yaml`
- **File Size:** 2257 bytes
- **Last Modified:** Tue Apr 22 2025 10:28:27 GMT-0500 (Peru Standard Time)
- **Last Commit:** update the code for the new name

```
apiVersion: apps/v1
kind: Deployment
metadata:
  name: ufl-rabbitmq-publisher
  #namespace: qa
  labels:
    app: ufl-rabbitmq-publisher
spec:
  replicas: 1
  selector:
    matchLabels:
      app: ufl-rabbitmq-publisher
  template:
    metadata:
      labels:
        app: ufl-rabbitmq-publisher
    spec:
      containers:
      - name: ufl-rabbitmq-publisher
        image: 826737140156.dkr.ecr.us-east-1.amazonaws.com/laliga-matchfantasy-rabbitmq-publisher:latest
        # This setting makes nodes pull the docker image every time before starting the pod. This is useful when debugging,
        # but should be turned off in production.
        imagePullPolicy: IfNotPresent
        env:
          - name: DATABASE_NAME
            value: "fanclash"
          - name: DATABASE_USER
            valueFrom:
              secretKeyRef:
                name: db-creds
                key: DATABASE_USER
          - name: DATABASE_PASSWORD
            valueFrom:
              secretKeyRef:
                name: db-creds
                key: DATABASE_PASSWORD
          - name: DATABASE_HOST
            valueFrom:
              secretKeyRef:
                name: db-creds
                key: DATABASE_HOST
          - name: RMQ_HOST
            value: "rabbitmq-0.rabbitmq-headless.rabbitmq.svc.cluster.local"
          - name: RMQ_PORT
            value: "5672"
          - name: RMQ_VHOST
            value: "ufl"
          - name: RMQ_USER
            valueFrom:
              secretKeyRef:
                name: rabbitmq
                key: RMQ_USER
          - name: RMQ_PASSWORD
            valueFrom:
              secretKeyRef:
                name: rabbitmq
                key: RMQ_PASSWORD
          - name: RMQ_EXCHANGES
            value: "match_event,fcm,games,system,game_updates"
          - name: RMQ_FCM_EXCHANGE
            value: "fcm"
          - name: RMQ_FCM_PUSHER_QUEUE
            value: "fcm:pusher"
          - name: WORKER_COUNT
            value: "20"
          - name: PREFETCH_COUNT
            value: "100"
          - name: FCM_CREDENTIALS
            valueFrom:
              secretKeyRef:
                name: fcm-creds
                key: FCM_CREDENTIALS
        resources:
          requests:
            cpu: 10m
```

## File: `deploy/eks/deployment_qa.yaml`
- **File Size:** 2257 bytes
- **Last Modified:** Tue Apr 22 2025 10:28:27 GMT-0500 (Peru Standard Time)
- **Last Commit:** update the code for the new name

```
apiVersion: apps/v1
kind: Deployment
metadata:
  name: ufl-rabbitmq-publisher
  #namespace: qa
  labels:
    app: ufl-rabbitmq-publisher
spec:
  replicas: 1
  selector:
    matchLabels:
      app: ufl-rabbitmq-publisher
  template:
    metadata:
      labels:
        app: ufl-rabbitmq-publisher
    spec:
      containers:
      - name: ufl-rabbitmq-publisher
        image: 736790963086.dkr.ecr.us-east-1.amazonaws.com/laliga-matchfantasy-rabbitmq-publisher:latest
        # This setting makes nodes pull the docker image every time before starting the pod. This is useful when debugging,
        # but should be turned off in production.
        imagePullPolicy: IfNotPresent
        env:
          - name: DATABASE_NAME
            value: "fanclash"
          - name: DATABASE_USER
            valueFrom:
              secretKeyRef:
                name: db-creds
                key: DATABASE_USER
          - name: DATABASE_PASSWORD
            valueFrom:
              secretKeyRef:
                name: db-creds
                key: DATABASE_PASSWORD
          - name: DATABASE_HOST
            valueFrom:
              secretKeyRef:
                name: db-creds
                key: DATABASE_HOST
          - name: RMQ_HOST
            value: "rabbitmq-0.rabbitmq-headless.rabbitmq.svc.cluster.local"
          - name: RMQ_PORT
            value: "5672"
          - name: RMQ_VHOST
            value: "ufl"
          - name: RMQ_USER
            valueFrom:
              secretKeyRef:
                name: rabbitmq
                key: RMQ_USER
          - name: RMQ_PASSWORD
            valueFrom:
              secretKeyRef:
                name: rabbitmq
                key: RMQ_PASSWORD
          - name: RMQ_EXCHANGES
            value: "match_event,fcm,games,system,game_updates"
          - name: RMQ_FCM_EXCHANGE
            value: "fcm"
          - name: RMQ_FCM_PUSHER_QUEUE
            value: "fcm:pusher"
          - name: WORKER_COUNT
            value: "20"
          - name: PREFETCH_COUNT
            value: "100"
          - name: FCM_CREDENTIALS
            valueFrom:
              secretKeyRef:
                name: fcm-creds
                key: FCM_CREDENTIALS
        resources:
          requests:
            cpu: 10m
```

## File: `deploy/gce/cloudbuild_production.yaml`
- **File Size:** 909 bytes
- **Last Modified:** Tue Apr 22 2025 10:28:27 GMT-0500 (Peru Standard Time)
- **Last Commit:** Refactor code for improved performance and readability

```
steps:
  - name: 'gcr.io/cloud-builders/docker'
    args: ['build', '-t', 'gcr.io/$PROJECT_ID/ufl-rabbitmq-publisher:$BRANCH_NAME.$COMMIT_SHA', '.']
    timeout: 180s
  - name: 'gcr.io/cloud-builders/docker'
    args: ['push', 'gcr.io/$PROJECT_ID/ufl-rabbitmq-publisher:$BRANCH_NAME.$COMMIT_SHA']
  - name: 'ubuntu'
    args: ['sed', '-i', 's/CLOUD_SQL_HOST/10.54.32.3/g', 'deploy/k8s/deployment.yaml']
  - name: 'ubuntu'
    args: ['sed', '-i', 's/PROJECT_ID/$PROJECT_ID/g', 'deploy/k8s/deployment.yaml']
  - name: 'ubuntu'
    args: ['sed', '-i', 's/BUILD_VERSION/$BRANCH_NAME.$COMMIT_SHA/g', 'deploy/k8s/deployment.yaml']
  - name: 'ubuntu'
    args: ['sed', '-i', 's/NAMESPACE/production/g', 'deploy/k8s/deployment.yaml']
  - name: 'gcr.io/cloud-builders/kubectl'
    args: ['apply', '-f', 'deploy/k8s']
    env:
      - 'CLOUDSDK_COMPUTE_ZONE=europe-west1-b'
      - 'CLOUDSDK_CONTAINER_CLUSTER=fanclash'
```

## File: `deploy/gce/cloudbuild_staging.yaml`
- **File Size:** 906 bytes
- **Last Modified:** Tue Apr 22 2025 10:28:27 GMT-0500 (Peru Standard Time)
- **Last Commit:** Refactor code for improved performance and readability

```
steps:
  - name: 'gcr.io/cloud-builders/docker'
    args: ['build', '-t', 'gcr.io/$PROJECT_ID/ufl-rabbitmq-publisher:$BRANCH_NAME.$COMMIT_SHA', '.']
    timeout: 180s
  - name: 'gcr.io/cloud-builders/docker'
    args: ['push', 'gcr.io/$PROJECT_ID/ufl-rabbitmq-publisher:$BRANCH_NAME.$COMMIT_SHA']
  - name: 'ubuntu'
    args: ['sed', '-i', 's/CLOUD_SQL_HOST/10.54.32.3/g', 'deploy/k8s/deployment.yaml']
  - name: 'ubuntu'
    args: ['sed', '-i', 's/PROJECT_ID/$PROJECT_ID/g', 'deploy/k8s/deployment.yaml']
  - name: 'ubuntu'
    args: ['sed', '-i', 's/BUILD_VERSION/$BRANCH_NAME.$COMMIT_SHA/g', 'deploy/k8s/deployment.yaml']
  - name: 'ubuntu'
    args: ['sed', '-i', 's/NAMESPACE/staging/g', 'deploy/k8s/deployment.yaml']
  - name: 'gcr.io/cloud-builders/kubectl'
    args: ['apply', '-f', 'deploy/k8s']
    env:
      - 'CLOUDSDK_COMPUTE_ZONE=europe-west1-b'
      - 'CLOUDSDK_CONTAINER_CLUSTER=fanclash'
```

## File: `deploy/k8s/deployment.yaml`
- **File Size:** 1732 bytes
- **Last Modified:** Tue Apr 22 2025 10:28:27 GMT-0500 (Peru Standard Time)
- **Last Commit:** Refactor code for improved performance and readability

```
apiVersion: apps/v1
kind: Deployment
metadata:
  name: ufl-rabbitmq-publisher
  namespace: NAMESPACE
  labels:
    app: ufl-rabbitmq-publisher
spec:
  replicas: 1
  selector:
    matchLabels:
      app: ufl-rabbitmq-publisher
  template:
    metadata:
      labels:
        app: ufl-rabbitmq-publisher
    spec:
      containers:
      - name: ufl-rabbitmq-publisher
        image: gcr.io/PROJECT_ID/ufl-rabbitmq-publisher:BUILD_VERSION
        # This setting makes nodes pull the docker image every time before
        # starting the pod. This is useful when debugging, but should be turned
        # off in production.
        imagePullPolicy: IfNotPresent
        env:
          - name: DATABASE_NAME
            valueFrom:
              configMapKeyRef:
                name: fanclash-config
                key: DATABASE_NAME
          - name: DATABASE_USER
            valueFrom:
              secretKeyRef:
                name: cloudsql
                key: username
          - name: DATABASE_PASSWORD
            valueFrom:
              secretKeyRef:
                name: cloudsql
                key: password
          - name: DATABASE_HOST
            value: "CLOUD_SQL_HOST"
          - name: RMQ_HOST
            value: rabbitmq-NAMESPACE
          - name: RMQ_PORT
            value: "5672"
          - name: RMQ_VHOST
            value: "ufl"
          - name: RMQ_USER
            value: "user"
          - name: RMQ_PASSWORD
            valueFrom:
              secretKeyRef:
                name: rabbitmq-NAMESPACE
                key: rabbitmq-password
          - name: RMQ_EXCHANGES
            value: "match_event,fcm,games,system,game_updates"
        resources:
          requests:
            cpu: 10m

```

## File: `images/laliga-matchfantasy-rabbitmq-publisher/deployment_dev.yaml`
- **File Size:** 1241 bytes
- **Last Modified:** Tue Apr 22 2025 10:28:27 GMT-0500 (Peru Standard Time)
- **Last Commit:** exclude the logs of the rabbitmq

```
apiVersion: apps/v1
kind: Deployment
metadata:
  name: laliga-matchfantasy-rabbitmq-publisher
  namespace: "fanclash-dev"
  labels:
    app: laliga-matchfantasy-rabbitmq-publisher
spec:
  replicas: 1
  selector:
    matchLabels:
      app: laliga-matchfantasy-rabbitmq-publisher
  template:
    metadata:
      annotations:
        ad.datadoghq.com/exclude: "true"
      labels:
        app: laliga-matchfantasy-rabbitmq-publisher
    spec:
      containers:
      - name: laliga-matchfantasy-rabbitmq-publisher
        image: registry.digitalocean.com/gameon-ams3/laliga-matchfantasy-rabbitmq-publisher:TAG_PLACEHOLDER
        # This setting makes nodes pull the docker image every time before starting the pod. This is useful when debugging, 
        # but should be turned off in production.
        imagePullPolicy: IfNotPresent
        envFrom:
          - configMapRef:
              name: fanclash-config
          - configMapRef:
              name: rmq-config
          - secretRef:
              name: db-creds
          - secretRef:
              name: rmq-creds
        env:
          - name: RMQ_EXCHANGES
            value: "match_event,fcm,games,system,game_updates"
        resources:
          requests:
            cpu: 10m
```

## File: `images/laliga-matchfantasy-rabbitmq-publisher/deployment_prd.yaml`
- **File Size:** 1241 bytes
- **Last Modified:** Tue Apr 22 2025 10:28:27 GMT-0500 (Peru Standard Time)
- **Last Commit:** add prd yaml file to main

```
apiVersion: apps/v1
kind: Deployment
metadata:
  name: laliga-matchfantasy-rabbitmq-publisher
  namespace: "prd-fanclash"
  labels:
    app: laliga-matchfantasy-rabbitmq-publisher
spec:
  replicas: 1
  selector:
    matchLabels:
      app: laliga-matchfantasy-rabbitmq-publisher
  template:
    metadata:
      annotations:
        ad.datadoghq.com/exclude: "true"
      labels:
        app: laliga-matchfantasy-rabbitmq-publisher
    spec:
      containers:
      - name: laliga-matchfantasy-rabbitmq-publisher
        image: registry.digitalocean.com/gameon-ams3/laliga-matchfantasy-rabbitmq-publisher:TAG_PLACEHOLDER
        # This setting makes nodes pull the docker image every time before starting the pod. This is useful when debugging, 
        # but should be turned off in production.
        imagePullPolicy: IfNotPresent
        envFrom:
          - configMapRef:
              name: fanclash-config
          - configMapRef:
              name: rmq-config
          - secretRef:
              name: db-creds
          - secretRef:
              name: rmq-creds
        env:
          - name: RMQ_EXCHANGES
            value: "match_event,fcm,games,system,game_updates"
        resources:
          requests:
            cpu: 10m
```

## File: `publisher/connector.go`
- **File Size:** 2047 bytes
- **Last Modified:** Tue Apr 22 2025 10:28:27 GMT-0500 (Peru Standard Time)
- **Last Commit:** Refactor code for improved performance and readability

```
package publisher

import (
	"context"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
)

const (
	ReconnectTimeout = time.Second * 3
)

type Connector struct {
	ctx       context.Context
	closeFunc func()
}

func (ec *Connector) Init() {
	ec.ctx, ec.closeFunc = context.WithCancel(context.Background())
}

func (ec *Connector) Close() {
	ec.closeFunc()
}

// redial continually connects to the URL, exiting the program when no longer possible
func (ec *Connector) connectToExchange(url string, exchanges []string) chan chan session {
	sessions := make(chan chan session)

	go func() {
		defer close(sessions)
		for {
			ctxAlive := ec.initSession(sessions, url, exchanges)
			if !ctxAlive {
				break
			}

			time.Sleep(ReconnectTimeout)
		}
	}()

	return sessions
}

func (ec *Connector) initSession(sessions chan chan session, url string, exchanges []string) bool {
	shouldCloseSess := true
	sess := make(chan session)
	defer func() {
		// session should be closed here only if there is no messages in channel
		// because if message exists in channel, end client will receive it and close sess by
		if shouldCloseSess {
			close(sess)
		}
	}()

	logrus.Info("trying to insert sess into sessions")
	select {
	case sessions <- sess:
	case <-ec.ctx.Done():
		logrus.Info("shutting down rabbit session factory")
		return false
	}

	logrus.WithField("url", url).Info("dial rabbitmq")
	conn, err := amqp.Dial(url)
	if err != nil {
		logrus.WithError(err).WithField("url", url).Error("Cannot (re)dial")
		return true
	}

	ch, err := conn.Channel()
	if err != nil {
		logrus.WithError(err).Error("cannot create channel")
		return true
	}

	for _, exchange := range exchanges {
		if err := ch.ExchangeDeclare(exchange, "fanout", true, false, false, false, nil); err != nil {
			logrus.WithError(err).Error("cannot declare fanout exchange")
			return true
		}
	}

	select {
	case sess <- session{conn, ch}:
	case <-ec.ctx.Done():
		logrus.Info("shutting down new session")
		return false
	}

	shouldCloseSess = false
	return true
}

```

## File: `publisher/event.go`
- **File Size:** 110 bytes
- **Last Modified:** Tue Apr 22 2025 10:28:27 GMT-0500 (Peru Standard Time)
- **Last Commit:** Refactor code for improved performance and readability

```
package publisher

type AMQPEvent struct {
	ID       int
	Exchange string
	Type     string
	Data     string
}

```

## File: `publisher/model.go`
- **File Size:** 676 bytes
- **Last Modified:** Tue Apr 22 2025 10:28:27 GMT-0500 (Peru Standard Time)
- **Last Commit:** Refactor code for improved performance and readability

```
package publisher

import (
	"time"

	"github.com/streadway/amqp"
)

// session composes an amqp.Connection with an amqp.Channel
type session struct {
	*amqp.Connection
	*amqp.Channel
}

// Close tears the connection down, taking the channel with it.
func (s session) Close() error {
	if s.Connection == nil {
		return nil
	}
	return s.Connection.Close()
}

type Message struct {
	Exchange    string
	Type        string
	Body        []byte
	Sub         session
	DeliveryTag uint64
}

func (m *Message) Success() {
	m.Sub.Ack(m.DeliveryTag, false)
}

func (m *Message) Fail() {
	// default redelivery time
	time.Sleep(time.Second * 3)
	m.Sub.Nack(m.DeliveryTag, false, true)
}

```

## File: `publisher/publisher.go`
- **File Size:** 3955 bytes
- **Last Modified:** Tue Apr 22 2025 10:28:27 GMT-0500 (Peru Standard Time)
- **Last Commit:** reduce to 3 sec, since listeners seems to not work

```
package publisher

import (
	"database/sql"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/lib/pq"
	"github.com/streadway/amqp"
)

const (
	NotifyTimeout = 3 * time.Second
)

const (
	getEventsSql = `
		select id,
		       exchange,
		       type,
		       data
		  from amqp_events
		 order by id
		 limit 100`

	deleteEventsSql = `
		delete
		  from amqp_events
		 where id = ANY($1)`
)

type EventPublisher struct {
	Connector
	db       *sql.DB
	listener *pq.Listener

	url       string
	exchanges []string
	queue     string
	messages  chan *Message
}

func (ev *EventPublisher) Start() {
	ev.messages = make(chan *Message)
	go func() {
		ev.startPublishing(ev.connectToExchange(ev.url, ev.exchanges))
		defer close(ev.messages)
	}()
}

func (ev *EventPublisher) Stop() {
	ev.Close()
}

// subscribe consumes deliveries from an exclusive queue from a fanout exchange and sends to the application specific messages chan.
func (ev *EventPublisher) startPublishing(sessions chan chan session) {
	for session := range sessions {
		pub, alive := <-session

		if !alive {
			continue
		}

		for {
			err := ev.processMessage(pub)
			if err != nil {
				break
			}
		}

		close(session)
	}
}

func (ev *EventPublisher) runInsideTransaction(fn func(tx *sql.Tx) error) error {
	tx, err := ev.db.Begin()
	if err != nil {
		logrus.WithError(err).Error("cannot start db transaction")
		return err
	}

	// Rollback the transaction on panics in the action. Don't swallow the
	// panic, though, let it propagate.
	defer func(tx *sql.Tx) {
		if r := recover(); r != nil {
			_ = tx.Rollback()
			panic(r)
		}
	}(tx)

	if err := fn(tx); err != nil {
		_ = tx.Rollback()
		return err
	}

	return tx.Commit()
}

func (ev *EventPublisher) processMessage(sess session) error {
	var (
		evID       int
		evExchange string
		evType     string
		evData     string
	)

	rows, err := ev.db.Query(getEventsSql)
	if err != nil {
		logrus.WithError(err).Error("error during query row")

		return nil
	}
	defer func() {
		_ = rows.Close()
	}()

	var processedEvents []int
	for rows.Next() {
		if err = rows.Scan(&evID, &evExchange, &evType, &evData); err != nil {
			logrus.WithError(err).Error("error during scan row")
			return err
		}

		// publish event to amqp
		err = ev.publishEvent(
			sess, &AMQPEvent{
				ID:       evID,
				Exchange: evExchange,
				Type:     evType,
				Data:     evData,
			},
		)

		if err != nil {
			logrus.WithError(err).Error("error during publish event to rabbitmq")
			return err
		}

		processedEvents = append(processedEvents, evID)
	}

	// get any error encountered during iteration
	err = rows.Err()
	if err != nil {
		return err
	}

	// wait a bit for next events
	if len(processedEvents) == 0 {
		// by default we should catch notify event and start next processing step
		// if event is not emitted during "NotifyTimeout" then go to next processing step immediately
		select {
		case <-time.After(NotifyTimeout):
			break
		case <-ev.listener.Notify:
			break
		}
	} else {
		_, err = ev.db.Exec(deleteEventsSql, pq.Array(processedEvents))
		if err != nil {
			logrus.WithError(err).Error("cannot delete processed amqp_events")
			return err
		}
	}

	return nil
}

func (ev *EventPublisher) publishEvent(sess session, event *AMQPEvent) error {
	// try to publish into channel
	logrus.WithField("id", event.ID).Info("publish event")

	err := sess.Publish(
		event.Exchange, "", false, false, amqp.Publishing{
			// use persistent delivery mode (mode = 2) for maximum consistency
			DeliveryMode: 2,
			Type:         event.Type,
			Body:         []byte(event.Data),
		},
	)
	if err != nil {
		logrus.WithError(err).Error("cannot publish")
		return err
	}

	return nil
}

func NewEventPublisher(url string, exchanges []string, db *sql.DB, listener *pq.Listener) *EventPublisher {
	publisher := &EventPublisher{
		db:        db,
		listener:  listener,
		url:       url,
		exchanges: exchanges,
	}
	publisher.Init()
	return publisher
}

```
