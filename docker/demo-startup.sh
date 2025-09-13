#!/bin/sh

echo "🚀 Starting FlexFlag Demo..."

# Set demo environment variables
export FLEXFLAG_DATABASE_HOST=${DATABASE_HOST:-localhost}
export FLEXFLAG_DATABASE_PORT=${DATABASE_PORT:-5432}
export FLEXFLAG_DATABASE_USERNAME=${DATABASE_USERNAME:-flexflag}
export FLEXFLAG_DATABASE_PASSWORD=${DATABASE_PASSWORD:-demo_password}
export FLEXFLAG_DATABASE_NAME=${DATABASE_NAME:-flexflag_demo}
export FLEXFLAG_SERVER_PORT=${PORT:-8080}
export FLEXFLAG_JWT_SECRET=${JWT_SECRET:-demo_jwt_secret_change_in_production}

# Demo mode settings
export FLEXFLAG_DEMO_MODE=true
export FLEXFLAG_DEMO_RESET_INTERVAL=1h
export FLEXFLAG_DEMO_MAX_FLAGS=50
export FLEXFLAG_DEMO_MAX_PROJECTS=5

echo "📊 Demo Settings:"
echo "  - Max Projects: $FLEXFLAG_DEMO_MAX_PROJECTS"
echo "  - Max Flags per Project: $FLEXFLAG_DEMO_MAX_FLAGS"
echo "  - Demo reset every: $FLEXFLAG_DEMO_RESET_INTERVAL"

# Wait for database to be ready
echo "⏳ Waiting for database..."
until pg_isready -h $FLEXFLAG_DATABASE_HOST -p $FLEXFLAG_DATABASE_PORT -U $FLEXFLAG_DATABASE_USERNAME; do
  sleep 1
done

# Run migrations
echo "🔄 Running database migrations..."
./bin/migrator up

# Insert demo data
echo "📝 Setting up demo data..."
if [ -f "./demo-data.sql" ]; then
  PGPASSWORD=$FLEXFLAG_DATABASE_PASSWORD psql -h $FLEXFLAG_DATABASE_HOST -p $FLEXFLAG_DATABASE_PORT -U $FLEXFLAG_DATABASE_USERNAME -d $FLEXFLAG_DATABASE_NAME -f ./demo-data.sql
fi

echo "🌟 Starting FlexFlag server..."
exec ./bin/server