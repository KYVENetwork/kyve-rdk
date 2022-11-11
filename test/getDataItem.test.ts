import { IRuntime, Node } from '@kyve/core-beta';
import Runtime from '../src/runtime';

import poolConfig from '../conf/pool.json';
import runtimeConfig from '../conf/runtime.json';

/*

TEST CASES - cache tests

* test

*/

describe('runtime tests', () => {
  let core: Node;
  let runtime: IRuntime;

  beforeEach(() => {
    runtime = new Runtime();
    core = new Node(runtime);
  });

  test('collect data item of start key', async () => {
    // ASSERT
    const getDataItemSpy = jest.spyOn(runtime, 'getDataItem');

    // ACT
    for (let source of runtimeConfig.sources) {
      await runtime.getDataItem(core, source, poolConfig.start_key);
    }

    // ASSERT
    expect(getDataItemSpy).toHaveBeenCalledTimes(1);
    expect(getDataItemSpy).not.toThrow();
  });

  test('get the next keys for first data bundle', async () => {
    // ASSERT
    const nextKeySpy = jest.spyOn(runtime, 'nextKey');

    let currentKey = poolConfig.start_key;

    // ACT
    for (let i = 0; i < parseInt(poolConfig.max_bundle_size); i++) {
      currentKey = await runtime.nextKey(currentKey);
    }

    // ASSERT
    expect(nextKeySpy).toHaveBeenCalledTimes(
      parseInt(poolConfig.max_bundle_size)
    );
    expect(nextKeySpy).not.toThrow();
  });

  test('get the next data items for first data bundle', async () => {
    // ASSERT
    const getDataItemSpy = jest.spyOn(runtime, 'getDataItem');
    const nextKeySpy = jest.spyOn(runtime, 'nextKey');

    let currentKey = poolConfig.start_key;

    // ACT
    for (let i = 0; i < parseInt(poolConfig.max_bundle_size); i++) {
      for (let source of runtimeConfig.sources) {
        await runtime.getDataItem(core, source, currentKey);
        currentKey = await runtime.nextKey(currentKey);
      }
    }

    // ASSERT
    expect(getDataItemSpy).toHaveBeenCalledTimes(
      parseInt(poolConfig.max_bundle_size)
    );
    expect(getDataItemSpy).not.toThrow();

    expect(nextKeySpy).toHaveBeenCalledTimes(
      parseInt(poolConfig.max_bundle_size)
    );
    expect(nextKeySpy).not.toThrow();
  });

  test('get the next data items for first data bundle including transformation', async () => {
    // ASSERT
    const getDataItemSpy = jest.spyOn(runtime, 'getDataItem');
    const transformDataItemSpy = jest.spyOn(runtime, 'transformDataItem');
    const nextKeySpy = jest.spyOn(runtime, 'nextKey');

    let currentKey = poolConfig.start_key;

    // ACT
    for (let i = 0; i < parseInt(poolConfig.max_bundle_size); i++) {
      for (let source of runtimeConfig.sources) {
        const dataItem = await runtime.getDataItem(core, source, currentKey);
        await runtime.transformDataItem(dataItem);
        currentKey = await runtime.nextKey(currentKey);
      }
    }

    // ASSERT
    expect(getDataItemSpy).toHaveBeenCalledTimes(
      parseInt(poolConfig.max_bundle_size)
    );
    expect(getDataItemSpy).not.toThrow();

    expect(transformDataItemSpy).toHaveBeenCalledTimes(
      parseInt(poolConfig.max_bundle_size)
    );
    expect(transformDataItemSpy).not.toThrow();

    expect(nextKeySpy).toHaveBeenCalledTimes(
      parseInt(poolConfig.max_bundle_size)
    );
    expect(nextKeySpy).not.toThrow();
  });

  test('create bundle summary for first data bundle', async () => {
    // ASSERT
    const getDataItemSpy = jest.spyOn(runtime, 'getDataItem');
    const transformDataItemSpy = jest.spyOn(runtime, 'transformDataItem');
    const nextKeySpy = jest.spyOn(runtime, 'nextKey');
    const summarizeDataBundleSpy = jest.spyOn(runtime, 'summarizeDataBundle');

    let currentKey = poolConfig.start_key;
    let bundle = [];

    // ACT
    for (let i = 0; i < parseInt(poolConfig.max_bundle_size); i++) {
      for (let source of runtimeConfig.sources) {
        const dataItem = await runtime.getDataItem(core, source, currentKey);
        const transformedDataItem = await runtime.transformDataItem(dataItem);

        bundle.push(transformedDataItem);

        currentKey = await runtime.nextKey(currentKey);
      }
    }

    await runtime.summarizeDataBundle(bundle);

    // ASSERT
    expect(getDataItemSpy).toHaveBeenCalledTimes(
      parseInt(poolConfig.max_bundle_size)
    );
    expect(getDataItemSpy).not.toThrow();

    expect(transformDataItemSpy).toHaveBeenCalledTimes(
      parseInt(poolConfig.max_bundle_size)
    );
    expect(transformDataItemSpy).not.toThrow();

    expect(nextKeySpy).toHaveBeenCalledTimes(
      parseInt(poolConfig.max_bundle_size)
    );
    expect(nextKeySpy).not.toThrow();

    expect(summarizeDataBundleSpy).toHaveBeenCalledTimes(1);
    expect(summarizeDataBundleSpy).not.toThrow();
  });
});
