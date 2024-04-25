import { ICacheProvider, Core } from "../../core";
import { JsonFileCache } from "./JsonFileCache";
import { MemoryCache } from "./MemoryCache";

/**
 * cacheProviderFactory creates the correct cache class
 * from the specified id. Current cache types are:
 *
 * 0 - JsonFile
 * x - Memory (default)
 *
 * @method cacheProviderFactory
 * @return {ICacheProvider}
 */
export function cacheProviderFactory(this: Core): ICacheProvider {
  switch (this.cache) {
    case "jsonfile":
      return new JsonFileCache();
    default:
      return new MemoryCache();
  }
}
