package mq

import (
	"Project1_Shop/models"
	"Project1_Shop/settings"
	"encoding/json"
	"fmt"
	"strconv"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
)

const (
	// OrderPendingQueue 订单待支付队列（30分钟 TTL，超时死信 → 取消队列）
	OrderPendingQueue         = "order.pending"
	OrderDeadLetterExchange   = "order.dlx"
	OrderDeadLetterRoutingKey = "order.expired"
	OrderExpiredQueue         = "order.expired"
	OrderPaymentQueue         = "order.payment"
	OrderPaymentRoutingKey    = "order.payment"
	OrderPaymentExchange      = "order.payment.exchange"
	SeckillQueue              = "seckill.order"
	SeckillExchange           = "seckill.exchange"
	SeckillRoutingKey         = "seckill.order"
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
	if err := ch.ExchangeDeclare(
		OrderPaymentExchange,
		"direct",
		true,
		false,
		false,
		false,
		nil,
	); err != nil {
		return fmt.Errorf("declare payment exchange failed: %w", err)
	}
	if _, err := ch.QueueDeclare(
		OrderPaymentQueue,
		true,
		false,
		false,
		false,
		nil,
	); err != nil {
		return fmt.Errorf("declare payment queue failed: %w", err)
	}
	if err := ch.QueueBind(OrderPaymentQueue, OrderPaymentRoutingKey, OrderPaymentExchange, false, nil); err != nil {
		return fmt.Errorf("bind payment queue failed: %w", err)
	}
	if err := ch.ExchangeDeclare(
		SeckillExchange,
		"direct",
		true, false, false, false, nil,
	); err != nil {
		return fmt.Errorf("declare seckill exchange failed: %w", err)
	}
	if _, err := ch.QueueDeclare(
		SeckillQueue,
		true, false, false, false, nil,
	); err != nil {
		return fmt.Errorf("declare seckill queue failed: %w", err)
	}
	if err := ch.QueueBind(SeckillQueue, SeckillRoutingKey, SeckillExchange, false, nil); err != nil {
		return fmt.Errorf("bind seckill queue failed: %w", err)
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

func PublishOrderPayment(orderID int64) error {
	return publish(OrderPaymentQueue, strconv.FormatInt(orderID, 10))
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

func PublishSeckillOrder(msg *models.SeckillMsg) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	ch, err := getChannel()
	if err != nil {
		return err
	}
	return ch.Publish(
		SeckillExchange,
		SeckillRoutingKey,
		false,
		false,
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  "application/json",
			Body:         data,
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
