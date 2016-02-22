# Kafka To Go

As promised, it's time to see Apache Kafka in action with a couple of microservices written in Go. You will be amazed how easy it's to have it up and running. For this showcase we'll fetch some Tweets and send them over the queue to a simple service that will print their content.

### Important To Know

First of all, you need to know that we are sending messages directly to Kafka broker, but we are reading them from ZooKeeper. Why is that? ZooKeeper will take care of remembering where did you finish reading, so that no message goes unread. 

We start by creating Kafka consumer and producer in a separate directory (package). Why should we do that? Chances are, that the environment will keep growing over time. If so, you should have the common stuff in a separate place. In our basic example it's definitely an overkill, but it would be a good practice any other time. We are going to use two cool repos: [Sirupsen's sarama](http://github.com/Shopify/sarama) and [wvanbergen's kafka](http://github.com/wvanbergen/kafka/consumergroup).

### Producer

We start by creating a simple wrapper over sarama's producer, so we need two functions. First, we need to construct a producer object:

    func NewProducer(broker string) (Producer, error) {
	    kafkaConfig := sarama.NewConfig()
	    kafkaConfig.Producer.Return.Successes = true

        kafkaProducer, err := sarama.NewSyncProducer([]string{broker}, kafkaConfig)
	    if err != nil {
		    return nil, err
	    }

	    return &producer{Producer: kafkaProducer}, nil
    }
    
and then a function for actually sending data over to Kafka:

    func NewProducer(broker string) (Producer, error) {
	    kafkaConfig := sarama.NewConfig()
	    kafkaConfig.Producer.Return.Successes = true

	kafkaProducer, err := sarama.NewSyncProducer([]string{broker}, kafkaConfig)
    	if err != nil {
	    	return nil, err
    	}

    	return &producer{Producer: kafkaProducer}, nil
    }

### Consumer
  
This is slightly more complicated, as we want to have our consumer to be a member of a consumer group (see previous post on Apache Kafka, of its documentation). Here, for better code clarity, we have three functions. First, again, is a constructor:

    func NewConsumer(consumerGroupName, topic, zkAddress string) (Consumer, error) {
    	c = Consumer{
	    	ConsumerGroupName: consumerGroupName,
		    ZkAddress:         zkAddress,
		    Topic:             topic,
	    }

	    return c, nil
    }

Then, we split our consuming into two parts: joining consumer group

    func (c *Consumer) Start() error {
        cg, err := consumergroup.JoinConsumerGroup(c.ConsumerGroupName, []string{c.Topic}, []string{c.ZkAddress}, nil)
        if err != nil {
            return err
        }
        defer cg.Close()

        runConsumer(c.Topic, cg)
        return nil
    }

and handling message callbacks:

    func runConsumer(topic string, provider messageProvider) {
        ...
        
        for {
            select {
            case msg := <-provider.Messages():
                log.Printf("[%s] Received message: %s", msg.Topic, string(msg.Value))

                if err := provider.CommitUpto(msg); err != nil {
                    log.Warnf("[%s] Consuming message: %v", msg.Topic, err)
                }
            case <-signals:
                return
            }
        }
    }

It's important to commit to the ZooKeeper the fact that the message is read, because by default it is nod done and you might find yourself reading the same message multiple times after restarts.

### Live Action

**Twitter service**:

    func main() {
        var (
            lastTweetID uint64
            handle      = "mycodesmells"
        )

        client := twitter.Connect()
        producer, err := kafka.NewProducer("0.0.0.0:9092")
        if err != nil {
            panic(fmt.Sprintf("Failed to initialize Kafka producer: %v", err))
        }

        for {
            select {
            case <-time.After(time.Second):
                tl, err := twitter.Tweets(client, handle, lastTweetID)
                if err != nil {
                    fmt.Printf("Failed to load Tweets from %s: %v", handle, err)
                }

                for _, t := range tl {
                    tweet := Tweet{Time: t.CreatedAt(), Text: t.Text()}
                    producer.SendMessage("example", tweet)
                }

                fmt.Printf("Found %d tweets\n", len(tl))

                if len(tl) > 0 {
                    lastTweetID = tl[0].Id()
                    tl[0].CreatedAt()
                }
            }
        }

    }

**Printer service**:

    func main() {
        consumer, err := kafka.NewConsumer("group", "example", "localhost:2181")
        if err != nil {
            panic(fmt.Sprintf("Failed to initialize Kafka consumer: %v", err))
        }

        err = consumer.Start()
        if err != nil {
            panic(fmt.Sprintf("Failed to start Kafka consumer: %v", err))
        }
    }

Now you can start the Twitter service, have it fetch a tweet list and shut it dow. The moment you start the second service, the messages are fetched from Kafka and processed.

A complete code for this example is available [on Github](https://github.com/mycodesmells/kafka-to-go).

