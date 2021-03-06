package main

import (
	"flag"
	"net/http"
	"os"
	"os/signal"

	"github.com/linkerd/linkerd2/controller/k8s"
	injector "github.com/linkerd/linkerd2/controller/proxy-injector"
	"github.com/linkerd/linkerd2/pkg/admin"
	"github.com/linkerd/linkerd2/pkg/flags"
	"github.com/linkerd/linkerd2/pkg/tls"
	log "github.com/sirupsen/logrus"
)

func main() {
	metricsAddr := flag.String("metrics-addr", ":9995", "address to serve scrapable metrics on")
	addr := flag.String("addr", ":8443", "address to serve on")
	kubeconfig := flag.String("kubeconfig", "", "path to kubeconfig")
	controllerNamespace := flag.String("controller-namespace", "linkerd", "namespace in which Linkerd is installed")
	webhookServiceName := flag.String("webhook-service", "linkerd-proxy-injector.linkerd.io", "name of the admission webhook")
	noInitContainer := flag.Bool("no-init-container", false, "whether to use an init container or the linkerd-cni plugin")
	flags.ConfigureAndParse()

	stop := make(chan os.Signal, 1)
	defer close(stop)
	signal.Notify(stop, os.Interrupt, os.Kill)

	k8sAPI, err := k8s.InitializeAPI(*kubeconfig, k8s.NS, k8s.RS)
	if err != nil {
		log.Fatalf("failed to initialize Kubernetes API: %s", err)
	}

	rootCA, err := tls.GenerateRootCAWithDefaults("Proxy Injector Mutating Webhook Admission Controller CA")
	if err != nil {
		log.Fatalf("failed to create root CA: %s", err)
	}

	webhookConfig, err := injector.NewWebhookConfig(k8sAPI, *controllerNamespace, *webhookServiceName, rootCA)
	if err != nil {
		log.Fatalf("failed to read the trust anchor file: %s", err)
	}

	mwc, err := webhookConfig.Create()
	if err != nil {
		log.Fatalf("failed to create the mutating webhook configurations resource: %s", err)
	}
	log.Infof("created mutating webhook configuration: %s", mwc.ObjectMeta.SelfLink)

	s, err := injector.NewWebhookServer(k8sAPI, *addr, *controllerNamespace, *noInitContainer, rootCA)
	if err != nil {
		log.Fatalf("failed to initialize the webhook server: %s", err)
	}

	k8sAPI.Sync()

	go func() {
		log.Infof("listening at %s", *addr)
		if err := s.ListenAndServeTLS("", ""); err != nil {
			if err == http.ErrServerClosed {
				return
			}
			log.Fatal(err)
		}
	}()
	go admin.StartServer(*metricsAddr)

	<-stop
	log.Info("shutting down webhook server")
	if err := s.Shutdown(); err != nil {
		log.Error(err)
	}
}
