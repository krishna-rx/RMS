CREATE TABLE IF NOT EXISTS users
(
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    email TEXT NOT NULL,
    password TEXT NOT NULL,
    archived_at TIMESTAMP WITH TIME ZONE
 );
CREATE TABLE IF NOT EXISTS roles(
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    role TEXT NOT NULL
);
CREATE TABLE IF NOT EXISTS user_roles(
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users (id) NOT NULL,
    roles_id UUID REFERENCES roles (id) NOT NULL,
    UNIQUE(user_id, roles_id)
);
CREATE TABLE IF NOT EXISTS addresses (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users (id) NOT NULL,
    address TEXT NOT NULL,
    latitude DOUBLE PRECISION NOT NULL,
    longitude DOUBLE PRECISION NOT NULL
);
CREATE TABLE IF NOT EXISTS restaurants (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    location TEXT NOT NULL,
    created_by UUID REFERENCES users(id) ON DELETE SET NULL,
    latitude DOUBLE PRECISION NOT NULL,
    longitude DOUBLE PRECISION NOT NULL
);
CREATE TABLE IF NOT EXISTS dishes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    description TEXT,
    price NUMERIC NOT NULL CHECK (price >=0),
    restaurant_id UUID REFERENCES restaurants(id),
    created_by UUID REFERENCES users(id) ON DELETE SET NULL
);
CREATE TABLE IF NOT EXISTS user_session (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    archived_at TIMESTAMP WITH TIME ZONE
)
