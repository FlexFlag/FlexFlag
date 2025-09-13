import { nodeResolve } from '@rollup/plugin-node-resolve';
import commonjs from '@rollup/plugin-commonjs';
import typescript from '@rollup/plugin-typescript';

const external = ['axios', 'eventemitter3', 'lru-cache', 'react', 'react/jsx-runtime', 'vue'];

export default [
  // CommonJS build
  {
    input: 'src/index.ts',
    output: {
      file: 'dist/index.js',
      format: 'cjs',
      sourcemap: true,
      exports: 'named'
    },
    external,
    plugins: [
      nodeResolve({
        preferBuiltins: false,
        browser: true
      }),
      commonjs(),
      typescript({
        declaration: false,
        declarationMap: false,
        emitDeclarationOnly: false
      })
    ]
  },
  
  // ES Module build
  {
    input: 'src/index.ts',
    output: {
      file: 'dist/index.esm.js',
      format: 'esm',
      sourcemap: true
    },
    external,
    plugins: [
      nodeResolve({
        preferBuiltins: false,
        browser: true
      }),
      commonjs(),
      typescript({
        declaration: false,
        declarationMap: false,
        emitDeclarationOnly: false
      })
    ]
  }
];