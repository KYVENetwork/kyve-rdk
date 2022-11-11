import { Node } from '@kyve/core-beta';
import MyCustomRuntime from './runtime';

const runtime = new MyCustomRuntime();

new Node(runtime).bootstrap();
