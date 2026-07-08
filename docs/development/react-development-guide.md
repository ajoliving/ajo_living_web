# React 开发规范

> 本文档及本仓库所有 .md 文件中均不得使用 emoji。
>
> 基于项目实际技术栈（React + TypeScript + Vite + React Router + Zustand + TanStack Query + Axios + shadcn/ui + Tailwind CSS），
> 参考现有 Vue 开发规范与社区最佳实践精简而来。只列规则，不做解释。

## 项目概况

- **技术栈**: React + TypeScript + Vite + React Router + Zustand + TanStack Query + Axios + shadcn/ui + Tailwind CSS
- **定位**: 设备管理与告警后台，包含 dashboard、设备、型号、告警、收件人等模块
- **路径别名**: `@/` 映射到 `src/`
- **ESLint**: 2 空格缩进
- **样式**: Tailwind CSS 为主，组件级补充 CSS Module 或局部样式

### 目录结构

```
src/
  api/               # API 请求函数（每个业务一个文件）
  components/        # 跨页面复用的通用组件
  features/
    <module>/
      components/    # 页面级子组件
      hooks/         # 页面业务逻辑 Hook
      pages/         # 路由页面
      schemas/       # 表单校验、查询参数校验
      types/         # 模块类型定义
  lib/               # axios、queryClient、utils 等基础能力
  routes/            # 路由定义与守卫
  stores/            # Zustand 全局状态
  types/             # 全局共享类型
  utils/             # 纯工具函数
```

---

## 一、命名

| 目标 | 规则 | 示例 |
| --- | --- | --- |
| 组件文件 | PascalCase 单词 | `DeviceEditDialog.tsx` |
| 页面文件 | PascalCase + `Page` | `DeviceListPage.tsx` |
| Hook 文件 | `use` + PascalCase | `useDeviceList.ts` |
| API 文件 | 小写 kebab-case | `device.ts` |
| Store 文件 | camelCase + `Store` | `authStore.ts` |
| 组件名 | PascalCase | `DeviceStatusCard` |
| prop | camelCase | `deviceId` |
| 事件函数 | `handle` + PascalCase | `handleSubmit` |
| CSS class | kebab-case 或 Tailwind 原子类 | `device-list` |
| 变量/函数 | camelCase | `isLoading`、`fetchDevices` |
| 常量 | SCREAMING_SNAKE | `MAX_RETRY_COUNT` |
| 类型/接口 | PascalCase | `Device`、`AlertRule` |
| 后端返回字段 | 与后端 JSON key 一致，不做无意义转换 | `deviceId`、`created_at` |

## 二、组件

### 2.1 文件顺序

```tsx
/*
 * 组件用途简要说明。
 * 1. 核心功能点。
 * 2. 核心功能点。
 * 3. 边界说明。
 */
import { useState } from 'react';

type Props = {
  visible: boolean;
};

export function DeviceEditDialog({ visible }: Props) {
  // hooks
  // state
  // memoized values
  // handlers
  // effects

  return <div />;
}
```

### 2.2 核心规则

- 统一使用函数组件，不使用 class component
- 一个文件默认只导出一个主组件
- 组件中只保留展示层和交互编排，复杂数据流程移到 Hook
- Props 必须显式声明类型，禁止隐式 `any`
- 页面组件只负责页面装配，业务状态放到 `hooks/`
- 可复用 2 次以上的 UI 片段再抽组件，不做过早碎片化
- 单个组件超过 200 行且职责可拆时，优先拆分子组件或 Hook

### 2.3 JSX 规则

- JSX 中不写复杂表达式，提取为变量、函数或 Hook
- 列表渲染必须使用稳定 `key`，禁止使用 `index` 作为可变列表 key
- 条件分支优先用早返回、三元表达式或拆组件，不堆叠多层 `&&`
- 多属性组件标签换行书写，闭合标签另起一行

## 三、TypeScript

### 3.1 类型优先级

```
interface > type > class
```

- 纯数据结构优先 `interface`
- 联合类型、映射类型、工具类型优先 `type`
- 只有在确实需要构造行为时才使用 `class`

### 3.2 禁止事项

- 禁止 `any`
- 禁止 `// @ts-ignore`
- 禁止隐式 `any`
- 禁止把后端响应先转成 `unknown` 再随意断言

### 3.3 实用规则

```ts
const [isLoading, setIsLoading] = useState<boolean>(false);

const handleSelect = (device: Device) => {
  // ...
};
```

## 四、Hooks

### 4.1 结构模板

```ts
/*
 * 设备列表页面业务 Hook。
 * 1. 查询与筛选。
 * 2. 分页控制。
 * 3. 创建、更新、删除动作。
 */
import { useQuery } from '@tanstack/react-query';

export function useDeviceList() {
  // state
  // query
  // mutations
  // handlers

  return {};
}
```

### 4.2 规则

- 页面业务逻辑统一收敛到 `features/<module>/hooks/`
- Hook 名称统一 `useXxx`
- 异步状态优先交给 TanStack Query 管理
- 副作用必须写在 `useEffect` 中，不在渲染阶段触发
- Hook 不直接操作 DOM，DOM 交互通过 `ref` 和组件内事件完成
- 错误提示在页面或组件层处理，不在基础 Hook 里散落 UI 提示

## 五、API 请求

### 5.1 axios 封装

- 统一实例在 `src/lib/http.ts`
- `baseURL` 通过 Vite 环境变量或代理配置，不在业务代码中硬编码
- 请求拦截器统一注入认证信息
- 响应拦截器统一处理认证失效和通用错误

### 5.2 函数风格

```ts
/*
 * 设备相关 API。
 * 1. 设备列表与详情。
 * 2. 设备创建、更新、删除。
 */
import { http } from '@/lib/http';
import type { Device } from '@/types/device';

export function getDevices(params?: Record<string, unknown>) {
  return http.get<Device[]>('/api/devices', { params });
}

export function createDevice(data: Device) {
  return http.post('/api/devices', data);
}
```

### 5.3 规则

- 每个资源一个文件
- API 层只负责请求，不负责弹窗、toast、跳转
- 路径命名与后端保持一致，不做二次包装
- 请求参数和返回值类型必须显式声明

## 六、路由

```ts
export const routes = [
  {
    path: '/devices',
    element: <DeviceListPage />,
    meta: { requiresAuth: true },
  },
];
```

- 路由定义统一放在 `src/routes/`
- 需要鉴权的页面通过路由元信息或守卫处理
- 不在页面组件内部重复做登录跳转判断
- 页面级懒加载通过 `lazy()` 和路由层统一处理

## 七、状态管理

- Zustand 只存全局共享状态，如登录态、用户信息、全局筛选条件
- 页面局部状态优先放组件或 Hook 内部
- store 中不得混入页面专属展示逻辑
- action 可异步，但必须显式处理失败分支

## 八、样式

- 样式优先使用 Tailwind CSS 原子类
- 组件存在复杂样式时，可补充 CSS Module
- 禁止在业务组件中堆积大段全局样式
- 不滥用 `!important`
- 主题变量、颜色、圆角、阴影统一收敛到设计令牌

## 九、页面组织

```
features/<module>/
  components/        # 模块内复用组件
  hooks/             # 模块业务 Hook
  pages/             # 路由页面
  schemas/           # zod 等校验定义
  types/             # 模块类型
```

- 一个 `features/<module>` 对应一个业务域
- 页面入口必须放在 `pages/`
- 模块内私有类型优先放 `types/`
- 跨模块共享类型放 `src/types/`

## 十、目录约束

- `components/` 只放跨页面复用组件
- `features/` 只放业务模块代码
- `lib/` 只放基础设施能力，不放业务规则
- `utils/` 只放纯函数，不放带副作用的业务逻辑
- `stores/` 只放全局状态，不放页面局部数据流
