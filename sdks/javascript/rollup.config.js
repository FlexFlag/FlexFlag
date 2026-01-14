import typescript from '@rollup/plugin-typescript';
import { nodeResolve } from '@rollup/plugin-node-resolve';
import commonjs from '@rollup/plugin-commonjs';
import dts from 'rollup-plugin-dts';

const external = ['axios', 'eventemitter3', 'lru-cache', 'react', 'react/jsx-runtime', 'vue'];

const createBuildConfig = (input, outputName) => [
  // CommonJS build
  {
    input,
    output: {
      file: `dist/${outputName}.js`,
      format: 'cjs',
      sourcemap: true,
      exports: 'named'
    },
    external,
    plugins: [
      nodeResolve({
        browser: true,
        preferBuiltins: false
      }),
      commonjs(),
      typescript({
        tsconfig: './tsconfig.json',
        sourceMap: true,
        declaration: false
      })
    ]
  },
  // ESM build
  {
    input,
    output: {
      file: `dist/${outputName}.esm.js`,
      format: 'esm',
      sourcemap: true
    },
    external,
    plugins: [
      nodeResolve({
        browser: true,
        preferBuiltins: false
      }),
      commonjs(),
      typescript({
        tsconfig: './tsconfig.json',
        sourceMap: true,
        declaration: false
      })
    ]
  },
  // Type definitions
  {
    input,
    output: {
      file: `dist/${outputName}.d.ts`,
      format: 'es'
    },
    external,
    plugins: [dts()]
  }
];

export default [
  // Core SDK (no React/Vue)
  ...createBuildConfig('src/index.ts', 'index'),
  // React integration
  ...createBuildConfig('src/index-react.ts', 'react'),
  // Vue integration
  ...createBuildConfig('src/index-vue.ts', 'vue')
];