import * as grpc from "@grpc/grpc-js";

export interface RuntimeConfig {
  host: string;
  port: number;
  channelOverride: grpc.Channel | undefined;
}
