# Vue 开发规范

> 本文档及本仓库所有 .md 文件中均不得使用 emoji。
>
> 基于项目实际技术栈（Vue 3 + TypeScript + Vite + Pinia + Ant Design Vue 4 + Axios），
> 参考 Vue 官方风格指南和社区最佳实践精简而来。只列规则，不做解释。

## 项目概况

- **技术栈**: Vue 3 + TypeScript + Vite + Pinia + Ant Design Vue 4 + Axios
- **定位**: 大屏设备管理后台，包含设备、楼宇、广告、公告、文件、版本、API Key、管理员等模块
- **路径别名**: `@/` 映射到 `src/`
- **ESLint**: 2 空格缩进
- **样式**: Sass scoped

### 目录结构

```
src/
  httpapis/          # API 请求函数（每个业务一个文件）
  model/             # TypeScript 接口/类定义
  pinia/             # Pinia 全局状态管理
  router/            # 路由定义 + 导航守卫
  stores/            # 额外状态（如 locations）
  utils/             # 工具函数
  components/        # 跨页面复用的通用组件
  views/
    <module>/
      <Module>View.vue          # 页面主组件
      use<Module>.ts            # 页面业务逻辑（Composable）
      components/               # 页面级子组件
```

---

## 一、命名

| 目标 | 规则 | 示例 |
| --- | --- | --- |
| 组件文件 | PascalCase 单词 | `DeviceEditDialog.vue` |
| composable 文件 | `use` + PascalCase | `useDevice.ts` |
| API 文件 | 小写 kebab-case | `device.ts` |
| 组件注册名 | PascalCase | `DeviceEditDialog` |
| prop/emit | camelCase | `deviceId`、`@update:visible` |
| template 引用 | PascalCase | `<DeviceEditDialog />` |
| 路由 name | PascalCase | `{ name: 'Device' }` |
| CSS class | kebab-case | `.device-list` |
| 变量/函数 | camelCase | `isLoading`、`fetchDevices` |
| 常量 | SCREAMING_SNAKE | `MAX_RETRY_COUNT` |
| 类型/接口 | PascalCase | `Device`、`DeviceSettings` |
| model 字段 | 与后端 JSON key 一致，不转换 | `deviceId`、`created_at` |

## 二、组件

### 2.1 单文件组件顺序

```vue
<!--
 * 组件用途简要说明
 * 1. 核心功能点
 * 2. 核心功能点
-->
<script setup lang="ts">
// imports
// props & emits
// composables
// reactive state
// computed
// methods
// lifecycle
</script>

<template>
  <!-- 结构 -->
</template>

<style scoped>
/* 样式 */
</style>
```

### 2.2 核心规则

- 统一使用 `<script setup lang="ts">`，不用 Options API
- 组件模板中只写简单表达式，复杂逻辑移到 script
- Props 必须 type 声明，不用运行时声明

```typescript
// 正确
const props = defineProps<{
  visible: boolean;
  mode: 'create' | 'edit';
  deviceData?: Device;
}>();

// 错误
const props = defineProps({
  visible: { type: Boolean, required: true },
});
```

- Emit 用类型声明

```typescript
const emit = defineEmits<{
  (e: 'update:visible', value: boolean): void;
  (e: 'created', data: Device): void;
}>();
```

### 2.3 组件拆分原则

- 重复使用 2 次以上 → 抽为组件
- 单个文件超过 200 行且逻辑可分 → 考虑拆分
- 弹窗、表单、列表项 → 天然适合拆为子组件
- 不要为了拆分而拆分，一个职责内聚的 300 行组件好过 5 个碎片组件

## 三、TypeScript

### 3.1 类型优先级

```
interface > type > class
```

- 纯数据结构用 `interface`
- 需要交叉/联合类型用 `type`
- 需要构造函数、默认值、静态方法用 `class`

### 3.2 禁止事项

- 禁止 `any`（必须用 `unknown` 或具体类型，特殊情况加注释）
- 禁止 `// @ts-ignore`（修复类型错误，而不是绕过）
- 禁止隐式 `any`（确保 `tsconfig.strict: true`）

### 3.3 实用规则

```typescript
// 1. ref 类型显式声明
const isLoading = ref<boolean>(false);

// 2. API 返回类型
const response = await getDevices(query); // 类型由 axiosInstance 泛型推导

// 3. 事件参数类型
const handleClick = (record: Device) => { ... };
```

## 四、Composable

### 4.1 结构模板

```typescript
/*
 * 模块名 - 业务逻辑
 * 1. 数据查询与分页
 * 2. 增删改操作
 * 3. 状态管理
 */
import { ref } from 'vue';
import { message } from 'ant-design-vue';
import { getXxxList, createXxx, updateXxx, deleteXxx } from '@/httpapis/xxx';
import { Xxx } from '@/model/xxxModel';

export const useXxxData = () => {
    // 1. 响应式状态
    const isLoading = ref(false);
    const data = ref<Xxx[]>([]);
    const pagination = ref({ currentPage: 1, pageSize: 10, total: 0 });
    const searchKeyword = ref('');

    // 2. 表格列定义
    const columns = [ ... ];

    // 3. 查询列表
    const list = async () => { ... };

    // 4. 获取详情
    const fetch = async (id: number) => { ... };

    // 5. 创建
    const create = async (form: Xxx) => { ... };

    // 6. 更新
    const update = async (form: Xxx) => { ... };

    // 7. 删除
    const remove = async (ids: number[]) => { ... };

    return {
        isLoading, data, pagination, searchKeyword, columns,
        list, fetch, create, update, remove,
    };
};
```

### 4.2 规则

- 文件放在对应 `views/<module>/` 下，不和视图分离
- 名称：`use<Module>Data` 或 `use<Module>`
- loading 态必须在 `finally` 中重置
- 错误用 `message.error()` 提示，不静默吞掉
- 不在 composable 里直接操作 DOM

## 五、API 请求

### 5.1 axios 封装

- 统一实例在 `httpapis/index.ts`
- baseURL 通过 vite proxy 配置，代码中不硬编码域名
- 请求拦截器注入 `Authorization` header
- 响应拦截器处理 401 → 清登录态 → 跳登录页

### 5.2 函数风格

```typescript
/*
 * 设备相关 API
 * 1. 设备 CRUD
 * 2. 轮播广告管理
 */
import axiosInstance from './index';
import { Device } from '@/model/deviceModel';

export const getDevices = (query?: object) => {
    return axiosInstance.get('/admin/device', { params: query });
};

export const createDevice = (data: Device) => {
    return axiosInstance.post('/admin/device', data);
};

export const updateDevice = (data: Device) => {
    return axiosInstance.put('/admin/device', data);
};

export const deleteDevices = (ids: number[]) => {
    return axiosInstance.delete('/admin/device', { data: { ids } });
};
```

### 5.3 规则

- 每个资源一个文件
- 函数不做错误处理，交给调用方决定如何处理
- 不在 API 层使用 `message.error()`
- 路径与后端一致，不额外封装

## 六、路由

```typescript
// router/routes.ts
const routes = [
    {
        name: 'Device',                        // PascalCase
        path: '/device',                       // kebab-case
        component: DeviceView,                 // 懒加载用 () => import(...)
        meta: { requiresAuth: true },
    },
];
```

- 需登录页面加 `meta: { requiresAuth: true }`
- 路由守卫集中在 `router/router.ts`
- 不在组件内做路由鉴权判断

## 七、Pinia

- 只存全局共享的状态（登录态、用户信息、全局配置）
- 模块级状态用 composable 内的 ref
- getter 做派生计算，不加副作用
- action 才允许异步操作

## 八、样式

- 组件样式用 `<style scoped lang="scss">`
- CSS class 命名 kebab-case
- UI 组件样式优先用 Ant Design Vue 的 props/slots/customize
- 不写全局样式覆盖（`main.ts` 中的 style.css 除外）
- 不使用 `!important`（Ant Design 深层覆盖用 `:deep()` 或全局 token）

## 九、模板

- 属性值始终带引号：`:visible="true"` 不写 `:visible=true`
- 多属性换行，闭合标签另起一行
- `v-for` 必须加 `:key`（用唯一 ID，不用 index）
- `v-if` 和 `v-for` 不在同一元素上
- 简单条件用 `v-if`/`v-else`，无序列表用 `v-show`
- 不在 template 中写复杂 JS 表达式，提取为 computed 或函数

```vue
<!-- 正确 -->
<a-table
  :columns="columns"
  :data-source="data"
  :loading="isLoading"
  :pagination="paginationConfig"
  @change="handleTableChange"
/>

<!-- 错误 -->
<a-table :columns="columns" :data-source="data" :loading="isLoading" :pagination="paginationConfig" @change="handleTableChange"/>
```

## 十、目录约束

```
views/<module>/          # 一个目录 = 一个路由页面
  <Module>View.vue       # 页面入口（必须有）
  use<Module>.ts         # 页面逻辑（必须有）
  components/            # 仅本页面用的组件

httpapis/                # API 函数，一个资源一个文件
model/                   # TS 类型定义，一个实体一个文件
pinia/                   # 全局状态
router/                  # 路由
utils/                   # 纯工具函数（与业务无关）
components/              # 跨页面复用的通用组件
```

## 附录：文件头部注释速查

```typescript
// .ts 文件
/*
 * 模块名 - 文件职责
 * 1. 功能点一
 * 2. 功能点二
 * 3. 功能点三
 */

// .vue 文件
<!--
 * 组件名 - 组件用途
 * 1. 功能点一
 * 2. 功能点二
-->
```
