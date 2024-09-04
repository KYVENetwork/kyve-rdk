import * as grpc from "@grpc/grpc-js";

export interface ProtocolConfig {
  host: string;
  port: number;
  channelOverride: grpc.Channel | undefined;
}
