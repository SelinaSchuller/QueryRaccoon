import {
  AddConnection,
  Connect,
  Disconnect,
} from "$wailsjs/go/bindings/ConnectionService";

export type DriverType = "postgresql" | "mysql" | "sqlite" | "mssql";

export type ConnectionConfig = {
  Host: string;
  Port: number;
  User: string;
  Password: string;
  Database: string;
  DriverType: DriverType;
};

export async function addConnection(config: ConnectionConfig): Promise<string> {
  return AddConnection(config);
}

export async function connectDB(id: string): Promise<void> {
  return Connect(id);
}

export async function disconnectDB(id: string): Promise<void> {
  return Disconnect(id);
}
