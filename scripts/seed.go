//go:build ignore

// Seed script creates development data for local testing.
// Run with: make seed  OR  go run scripts/seed.go
//
// Credentials created:
//
//	Super Admin : phone=+910000000001  password=SuperAdmin@123
//	Tenant Admin: phone=+910000000002  password=TenantAdmin@123
//	User        : phone=+910000000003  password=User@1234
//
// NEVER use these credentials in production.
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"github.com/autobee/server/internal/config"
	"github.com/autobee/server/internal/database"
	"github.com/autobee/server/pkg/password"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config error: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	db, err := database.Connect(ctx, cfg.Mongo.URI, cfg.Mongo.Database)
	if err != nil {
		log.Fatalf("mongo connect error: %v", err)
	}
	defer db.Disconnect(ctx) //nolint:errcheck

	if err := database.EnsureIndexes(ctx, db.Database); err != nil {
		log.Fatalf("index creation error: %v", err)
	}

	// ─── Tenant ─────────────────────────────────────────────────────────────
	tenantID := upsertTenant(ctx, db.Database)
	fmt.Printf("✅ Tenant:       %s (id: %s)\n", "Autobee Dev Tenant", tenantID.Hex())

	// ─── Super Admin ─────────────────────────────────────────────────────────
	superAdminID := upsertUser(ctx, db.Database, seedUser{
		TenantID: tenantID,
		Name:     "Super Admin",
		Phone:    "+910000000001",
		Email:    "superadmin@autobee.com",
		Password: "SuperAdmin@123",
		Role:     "super_admin",
	})
	fmt.Printf("✅ Super Admin:  phone=+910000000001  password=SuperAdmin@123  (id: %s)\n", superAdminID.Hex())

	// ─── Tenant Admin ────────────────────────────────────────────────────────
	tenantAdminID := upsertUser(ctx, db.Database, seedUser{
		TenantID: tenantID,
		Name:     "Tenant Admin",
		Phone:    "+910000000002",
		Email:    "tenantadmin@autobee.com",
		Password: "TenantAdmin@123",
		Role:     "tenant_admin",
	})
	fmt.Printf("✅ Tenant Admin: phone=+910000000002  password=TenantAdmin@123 (id: %s)\n", tenantAdminID.Hex())

	// ─── Regular User ─────────────────────────────────────────────────────────
	userID := upsertUser(ctx, db.Database, seedUser{
		TenantID: tenantID,
		Name:     "Dev User",
		Phone:    "+910000000003",
		Email:    "user@autobee.com",
		Password: "User@1234",
		Role:     "user",
	})
	fmt.Printf("✅ User:         phone=+910000000003  password=User@1234       (id: %s)\n", userID.Hex())

	// ─── Service Center ───────────────────────────────────────────────────────
	scID := upsertServiceCenter(ctx, db.Database, tenantID)
	fmt.Printf("✅ Service Center: %s (id: %s)\n", "Dev Service Center", scID.Hex())

	// ─── Service ──────────────────────────────────────────────────────────────
	svcID := upsertService(ctx, db.Database, tenantID, scID)
	fmt.Printf("✅ Service: %s (id: %s)\n", "Oil Change", svcID.Hex())

	// ─── Vehicle ──────────────────────────────────────────────────────────────
	vID := upsertVehicle(ctx, db.Database, tenantID, userID)
	fmt.Printf("✅ Vehicle: %s (id: %s)\n", "KA01DE0001", vID.Hex())

	fmt.Println("\n🌱 Seed complete.")
	fmt.Printf("   Tenant ID: %s\n", tenantID.Hex())
	fmt.Println("   Copy this Tenant ID when testing signup.")
}

// ─── Helpers ────────────────────────────────────────────────────────────────

type seedUser struct {
	TenantID bson.ObjectID
	Name     string
	Phone    string
	Email    string
	Password string
	Role     string
}

func upsertTenant(ctx context.Context, db *mongo.Database) bson.ObjectID {
	col := db.Collection("tenants")
	filter := bson.D{{Key: "name", Value: "Autobee Dev Tenant"}}

	var existing bson.M
	if err := col.FindOne(ctx, filter).Decode(&existing); err == nil {
		return existing["_id"].(bson.ObjectID)
	}

	id := bson.NewObjectID()
	now := time.Now().UTC()
	doc := bson.D{
		{Key: "_id", Value: id},
		{Key: "name", Value: "Autobee Dev Tenant"},
		{Key: "status", Value: "active"},
		{Key: "createdAt", Value: now},
		{Key: "updatedAt", Value: now},
	}
	if _, err := col.InsertOne(ctx, doc); err != nil {
		log.Fatalf("insert tenant: %v", err)
	}
	return id
}

func upsertUser(ctx context.Context, db *mongo.Database, u seedUser) bson.ObjectID {
	col := db.Collection("users")
	filter := bson.D{
		{Key: "tenantId", Value: u.TenantID},
		{Key: "phone", Value: u.Phone},
	}

	var existing bson.M
	if err := col.FindOne(ctx, filter).Decode(&existing); err == nil {
		return existing["_id"].(bson.ObjectID)
	}

	hash, err := password.Hash(u.Password)
	if err != nil {
		log.Fatalf("hash password: %v", err)
	}

	id := bson.NewObjectID()
	now := time.Now().UTC()
	doc := bson.D{
		{Key: "_id", Value: id},
		{Key: "tenantId", Value: u.TenantID},
		{Key: "name", Value: u.Name},
		{Key: "phone", Value: u.Phone},
		{Key: "email", Value: u.Email},
		{Key: "passwordHash", Value: hash},
		{Key: "role", Value: u.Role},
		{Key: "status", Value: "active"},
		{Key: "phoneVerified", Value: true},
		{Key: "createdAt", Value: now},
		{Key: "updatedAt", Value: now},
	}

	opts := options.UpdateOne().SetUpsert(true)
	_, err = col.UpdateOne(ctx, filter, bson.D{{Key: "$setOnInsert", Value: doc}}, opts)
	if err != nil {
		log.Fatalf("upsert user: %v", err)
	}
	return id
}

func upsertServiceCenter(ctx context.Context, db *mongo.Database, tenantID bson.ObjectID) bson.ObjectID {
	col := db.Collection("service_centers")
	filter := bson.D{
		{Key: "tenantId", Value: tenantID},
		{Key: "name", Value: "Dev Service Center"},
	}

	var existing bson.M
	if err := col.FindOne(ctx, filter).Decode(&existing); err == nil {
		return existing["_id"].(bson.ObjectID)
	}

	id := bson.NewObjectID()
	now := time.Now().UTC()
	doc := bson.D{
		{Key: "_id", Value: id},
		{Key: "tenantId", Value: tenantID},
		{Key: "name", Value: "Dev Service Center"},
		{Key: "address", Value: "123 Dev Street, Bengaluru, Karnataka 560001"},
		{Key: "phone", Value: "+919999999999"},
		{Key: "isActive", Value: true},
		{Key: "createdAt", Value: now},
		{Key: "updatedAt", Value: now},
	}
	if _, err := col.InsertOne(ctx, doc); err != nil {
		log.Fatalf("insert service center: %v", err)
	}
	return id
}

func upsertService(ctx context.Context, db *mongo.Database, tenantID, scID bson.ObjectID) bson.ObjectID {
	col := db.Collection("services")
	filter := bson.D{
		{Key: "tenantId", Value: tenantID},
		{Key: "name", Value: "Oil Change"},
	}

	var existing bson.M
	if err := col.FindOne(ctx, filter).Decode(&existing); err == nil {
		return existing["_id"].(bson.ObjectID)
	}

	id := bson.NewObjectID()
	now := time.Now().UTC()
	doc := bson.D{
		{Key: "_id", Value: id},
		{Key: "tenantId", Value: tenantID},
		{Key: "serviceCenterId", Value: scID},
		{Key: "name", Value: "Oil Change"},
		{Key: "description", Value: "Full synthetic oil change with filter replacement"},
		{Key: "durationMinutes", Value: 30},
		{Key: "price", Value: 999.0},
		{Key: "isActive", Value: true},
		{Key: "createdAt", Value: now},
		{Key: "updatedAt", Value: now},
	}
	if _, err := col.InsertOne(ctx, doc); err != nil {
		log.Fatalf("insert service: %v", err)
	}
	return id
}

func upsertVehicle(ctx context.Context, db *mongo.Database, tenantID, userID bson.ObjectID) bson.ObjectID {
	col := db.Collection("vehicles")
	filter := bson.D{
		{Key: "tenantId", Value: tenantID},
		{Key: "registrationNumber", Value: "KA01DE0001"},
	}

	var existing bson.M
	if err := col.FindOne(ctx, filter).Decode(&existing); err == nil {
		return existing["_id"].(bson.ObjectID)
	}

	id := bson.NewObjectID()
	now := time.Now().UTC()
	doc := bson.D{
		{Key: "_id", Value: id},
		{Key: "tenantId", Value: tenantID},
		{Key: "userId", Value: userID},
		{Key: "registrationNumber", Value: "KA01DE0001"},
		{Key: "make", Value: "Maruti Suzuki"},
		{Key: "model", Value: "Swift"},
		{Key: "year", Value: 2022},
		{Key: "fuelType", Value: "petrol"},
		{Key: "createdAt", Value: now},
		{Key: "updatedAt", Value: now},
	}
	if _, err := col.InsertOne(ctx, doc); err != nil {
		log.Fatalf("insert vehicle: %v", err)
	}
	return id
}
