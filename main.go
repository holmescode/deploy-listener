package listener

import "log"

func main() {
	consumer, err := NewConsumer("amqp://localhost", "holmescode.deployments", "topic", "holmescode.deploymentsQueue", "", "deploy-listener")
	if err != nil {
		log.Fatalf("%s", err)
	}

	select {}

	if err := consumer.Shutdown(); err != nil {
		log.Fatalf("Could not gracefully shutdown: %s", err)
	}
}
