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

  test('getDataItem: collect data item of start key', async () => {
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

  test('nextKey: get the next 10 keys', async () => {
    // ASSERT
    const nextKeySpy = jest.spyOn(runtime, 'nextKey');

    let currentKey = poolConfig.start_key;

    // ACT
    for (let i = 0; i < 10; i++) {
      currentKey = await runtime.nextKey(currentKey);
    }

    // ASSERT
    expect(nextKeySpy).toHaveBeenCalledTimes(10);
    expect(nextKeySpy).not.toThrow();
  });

  test('getDataItem: collect first 10 data items', async () => {
    // ASSERT
    const getDataItemSpy = jest.spyOn(runtime, 'getDataItem');
    const nextKeySpy = jest.spyOn(runtime, 'nextKey');

    let currentKey = poolConfig.start_key;

    // ACT
    for (let i = 0; i < 10; i++) {
      for (let source of runtimeConfig.sources) {
        await runtime.getDataItem(core, source, currentKey);
        currentKey = await runtime.nextKey(currentKey);
      }
    }

    // ASSERT
    expect(getDataItemSpy).toHaveBeenCalledTimes(10);
    expect(getDataItemSpy).not.toThrow();

    expect(nextKeySpy).toHaveBeenCalledTimes(10);
    expect(nextKeySpy).not.toThrow();
  });
});
