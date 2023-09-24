package database

import (
	"context"
	"facebook-business-sdk-codegen-relationships/models"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

var Driver neo4j.DriverWithContext
var Session neo4j.SessionWithContext

func InitNeo4j() {
	ctx := context.Background()

	driver, err := neo4j.NewDriverWithContext(
		"bolt://localhost:7687",
		neo4j.NoAuth(),
	)

	Driver = driver

	err = Driver.VerifyConnectivity(ctx)
	if err != nil {
		panic(err)
	}

	Session = driver.NewSession(ctx, neo4j.SessionConfig{DatabaseName: "neo4j"})
}

func ResetDatabase() error {
	ctx := context.Background()

	_, err := Session.ExecuteWrite(ctx,
		func(tx neo4j.ManagedTransaction) (any, error) {
			_, err := tx.Run(
				ctx, `
					MATCH (n) DETACH DELETE n
				`, nil)

			return "", err
		})

	return err
}

func CreateAdObject(name string, fields []models.AdObjectField) error {
	ctx := context.Background()

	_, err := Session.ExecuteWrite(ctx,
		func(tx neo4j.ManagedTransaction) (any, error) {
			objId, err := addAddObject(ctx, tx, name)

			if err != nil {
				return "", err
			}

			for _, field := range fields {
				_ = addFieldToObject(ctx, tx, objId, field.Name, field.Type)
			}

			return objId, nil
		})

	return err
}

func CreateLinkBetweenFieldAndObject(fieldName string) error {
	ctx := context.Background()

	_, err := Session.ExecuteWrite(ctx,
		func(tx neo4j.ManagedTransaction) (any, error) {
			_, err := tx.Run(
				ctx, `
					MATCH (o:AdObject {name: $name})
					MATCH (f:AdObjectField {type: $name})
					MERGE (f)-[:IS_TYPE]->(o)
				`, map[string]any{
					"name": fieldName,
				})

			return "", err
		})

	return err
}

func addAddObject(ctx context.Context, tx neo4j.ManagedTransaction, name string) (string, error) {
	result, err := tx.Run(
		ctx, `
			CREATE (o:AdObject {id: randomuuid(), name: $name}) RETURN o.id AS id
		`, map[string]any{
			"name": name,
		})

	if err != nil {
		return "", err
	}

	obj, err := result.Single(ctx)
	if err != nil {
		return "", err
	}

	objId, _ := obj.AsMap()["id"]

	return objId.(string), err
}

func addFieldToObject(ctx context.Context, tx neo4j.ManagedTransaction, objId string, fieldName string, fieldType string) error {
	_, err := tx.Run(
		ctx, `
			MATCH (o:AdObject {id: $objId})
			CREATE (p:AdObjectField {name: $fieldName, type: $fieldType})
			MERGE (p)-[:IS_FIELD]->(o)
        `, map[string]any{
			"objId":     objId,
			"fieldName": fieldName,
			"fieldType": fieldType,
		})

	return err
}
