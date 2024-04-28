package entity

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

type DynamoInf struct {
	session   *dynamodb.Client
	tableName *string
}

func NewDynamoInf(session *dynamodb.Client, tableName string) *DynamoInf {
	di := new(DynamoInf)
	di.session = session
	di.tableName = aws.String(tableName)
	return di
}

func (di *DynamoInf) FindOne(key *expression.KeyConditionBuilder, res interface{}) (interface{}, error) {
	expr, _ := expression.NewBuilder().
		WithKeyCondition(*key).
		Build()

	input := &dynamodb.QueryInput{
		Limit:                     aws.Int32(1),
		TableName:                 di.tableName,
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		KeyConditionExpression:    expr.KeyCondition(),
		FilterExpression:          expr.Filter(),
	}

	response, err := di.session.Query(context.TODO(), input)

	if err != nil {
		return nil, err
	}

	err = attributevalue.UnmarshalMap(response.Items[0], res)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (di *DynamoInf) FindSelectAll(
	key *expression.KeyConditionBuilder,
	project *expression.ProjectionBuilder,
	res interface{},
) (interface{}, error) {
	resSlice, ok := res.([]interface{})
	if !ok {
		panic("res should be []interface{}")
	}

	expr, _ := expression.NewBuilder().
		WithKeyCondition(*key).
		WithProjection(*project).
		Build()

	input := dynamodb.QueryInput{
		TableName:                 di.tableName,
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		KeyConditionExpression:    expr.KeyCondition(),
		FilterExpression:          expr.Filter(),
	}

	paginator := dynamodb.NewQueryPaginator(di.session, &input)
	for paginator.HasMorePages() {
		out, err := paginator.NextPage(context.TODO())
		if err != nil {
			return nil, err
		}

		var pItems []interface{}
		err = attributevalue.UnmarshalListOfMaps(out.Items, &pItems)
		if err != nil {
			return nil, err
		}

		resSlice = append(resSlice, pItems...)
	}

	return resSlice, nil
}

func (di *DynamoInf) Upsert(item map[string]types.AttributeValue) (interface{}, error) {
	input := &dynamodb.PutItemInput{
		TableName: di.tableName,
		Item:      item,
	}

	return di.session.PutItem(context.TODO(), input)
}

func (di *DynamoInf) Update(
	key map[string]types.AttributeValue,
	condition *expression.ConditionBuilder,
	update *expression.UpdateBuilder,
) error {
	var builder = expression.NewBuilder().
		WithUpdate(*update)
	if condition != nil {
		builder = builder.WithCondition(*condition)
	}
	expr, _ := builder.Build()

	input := &dynamodb.UpdateItemInput{
		TableName:                 di.tableName,
		Key:                       key,
		UpdateExpression:          expr.Update(),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		ConditionExpression:       expr.Condition(),
	}

	_, err := di.session.UpdateItem(context.TODO(), input)
	return err
}

func (di *DynamoInf) Delete(
	key map[string]types.AttributeValue,
) error {
	input := &dynamodb.DeleteItemInput{
		TableName: di.tableName,
		Key:       key,
	}

	_, err := di.session.DeleteItem(context.TODO(), input)
	return err
}
