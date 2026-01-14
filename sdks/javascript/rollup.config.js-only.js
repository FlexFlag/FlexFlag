import { nodeResolve } from '@rollup/plugin-node-resolve';
import commonjs from '@rollup/plugin-commonjs';
import typescript from '@rollup/plugin-typescript';

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
    input,
    output: {
      file: `dist/${outputName}.esm.js`,
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

export default [
  // Core SDK (no React/Vue)
  ...createBuildConfig('src/index.ts', 'index'),
  // React integration
  ...createBuildConfig('src/index-react.ts', 'react'),
  // Vue integration
  ...createBuildConfig('src/index-vue.ts', 'vue')
];