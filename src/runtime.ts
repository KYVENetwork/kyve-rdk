import { DataItem, IRuntime, Node } from '@kyve/core-beta';
import { name, version } from '../package.json';

export default class MyCustomRuntime implements IRuntime {
  public name = name;
  public version = version;

  async getDataItem(
    core: Node,
    source: string,
    key: string
  ): Promise<DataItem> {
    return {
      key,
      value: null,
    };
  }

  async transformDataItem(item: DataItem): Promise<DataItem> {
    return item;
  }

  async validateDataItem(
    core: Node,
    proposedDataItem: DataItem,
    validationDataItem: DataItem
  ): Promise<boolean> {
    return true;
  }

  async summarizeDataBundle(bundle: DataItem[]): Promise<string> {
    return '';
  }

  async nextKey(key: string): Promise<string> {
    return key;
  }
}
