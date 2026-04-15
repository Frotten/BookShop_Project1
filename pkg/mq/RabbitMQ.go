package mq

import (
	"Project1_Shop/settings"
	"fmt"
	"strconv"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
)

const (
	// OrderPendingQueue 订单待支付队列（30分钟 TTL，超时死信 → 取消队列）
	OrderPendingQueue = "order.pending"
	// OrderDeadLetterExchange 死信交换机 & 路由键
	OrderDeadLetterExchange   = "order.dlx"
	OrderDeadLetterRoutingKey = "order.expired"
	// OrderExpiredQueue 死信消费队列（实际执行取消逻辑）
	OrderExpiredQueue = "order.expired"
)

var (
	globalConn *amqp.Connection
	globalCh   *amqp.Channel
	mu         sync.Mutex
)

func Init(cfg *settings.RabbitMQConfig) error {
	dsn := fmt.Sprintf("amqp://%s:%s@%s:%d/", cfg.User, cfg.Password, cfg.Host, cfg.Port)
	conn, err := amqp.Dial(dsn)
	if err != nil {
		return fmt.Errorf("rabbitmq dial failed: %w", err)
	}
	ch, err := conn.Channel()
	if err != nil {
		_ = conn.Close()
		return fmt.Errorf("rabbitmq open channel failed: %w", err)
	}
	if err := ch.ExchangeDeclare(
		OrderDeadLetterExchange,
		"direct",
		true,
		false,
		false,
		false,
		nil,
	); err != nil {
		return fmt.Errorf("declare dlx exchange failed: %w", err)
	}
	if _, err := ch.QueueDeclare(
		OrderExpiredQueue,
		true,
		false,
		false,
		false,
		nil,
	); err != nil {
		return fmt.Errorf("declare expired queue failed: %w", err)
	}
	if err := ch.QueueBind(OrderExpiredQueue, OrderDeadLetterRoutingKey, OrderDeadLetterExchange, false, nil); err != nil {
		return fmt.Errorf("bind expired queue failed: %w", err)
	}
	if _, err := ch.QueueDeclare(
		OrderPendingQueue,
		true,
		false,
		false,
		false,
		amqp.Table{
			"x-message-ttl":             int32(30 * 60 * 1000), //这里是毫秒
			"x-dead-letter-exchange":    OrderDeadLetterExchange,
			"x-dead-letter-routing-key": OrderDeadLetterRoutingKey,
		},
	); err != nil {
		return fmt.Errorf("declare pending queue failed: %w", err)
	}

	globalConn = conn
	globalCh = ch
	zap.L().Info("RabbitMQ init success",
		zap.String("host", cfg.Host),
		zap.Int("port", cfg.Port),
	)
	return nil
}

func Close() {
	mu.Lock()
	defer mu.Unlock()
	if globalCh != nil {
		_ = globalCh.Close()
	}
	if globalConn != nil {
		_ = globalConn.Close()
	}
}

func getChannel() (*amqp.Channel, error) {
	mu.Lock()
	defer mu.Unlock()
	if globalCh != nil && !globalCh.IsClosed() {
		return globalCh, nil
	}
	zap.L().Warn("RabbitMQ channel closed, reconnecting...")
	if globalConn == nil || globalConn.IsClosed() {
		return nil, fmt.Errorf("rabbitmq connection is closed")
	}
	ch, err := globalConn.Channel()
	if err != nil {
		return nil, err
	}
	globalCh = ch
	return globalCh, nil
}

func PublishOrderPending(orderID int64) error {
	return publish(OrderPendingQueue, strconv.FormatInt(orderID, 10))
}

func publish(queue, body string) error {
	ch, err := getChannel()
	if err != nil {
		return err
	}
	return ch.Publish(
		"", //不填，表示默认交换机
		queue,
		false,
		false,
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  "text/plain",
			Body:         []byte(body),
			Timestamp:    time.Now(),
		},
	)
}

type ConsumeFunc func(orderID int64) bool

func StartConsumer(queue string, handler ConsumeFunc) error {
	mu.Lock()
	if globalConn == nil || globalConn.IsClosed() {
		mu.Unlock()
		return fmt.Errorf("rabbitmq connection not initialized")
	}
	conn := globalConn
	mu.Unlock()
	ch, err := conn.Channel()
	if err != nil {
		return fmt.Errorf("open consumer channel failed: %w", err)
	}
	if err := ch.Qos(1, 0, false); err != nil {
		_ = ch.Close()
		return fmt.Errorf("set QoS failed: %w", err)
	}
	msgs, err := ch.Consume(
		queue,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		_ = ch.Close()
		return fmt.Errorf("consume queue %s failed: %w", queue, err)
	}
	go func() {
		defer func() {
			_ = ch.Close()
			zap.L().Info("RabbitMQ consumer stopped", zap.String("queue", queue))
		}()
		zap.L().Info("RabbitMQ consumer started", zap.String("queue", queue))
		for msg := range msgs {
			var orderID int64
			if _, err := fmt.Sscanf(string(msg.Body), "%d", &orderID); err != nil {
				zap.L().Error("parse orderID failed", zap.String("body", string(msg.Body)), zap.Error(err))
				_ = msg.Nack(false, false)
				continue
			}
			if ok := handler(orderID); ok {
				_ = msg.Ack(false)
			} else {
				_ = msg.Nack(false, false)
			}
		}
	}()
	return nil
}
