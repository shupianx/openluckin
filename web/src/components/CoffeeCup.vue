<script setup>
import { onMounted, onBeforeUnmount, ref } from 'vue'
import * as THREE from 'three'
import { RoomEnvironment } from 'three/addons/environments/RoomEnvironment.js'
import logoBandUrl from '../assets/luckin-logo.svg'
import logoDeerUrl from '../assets/luckin-deer.svg'

// 瑞幸品牌蓝（取自官方 logo 底色）
const BRAND_HEX = '#212370'
const BRAND = 0x212370

const container = ref(null)
let renderer, scene, camera, stage, raf, onResize, onMouseMove, pmrem, envTex

function loadImage(src) {
  return new Promise((resolve) => {
    const img = new Image()
    img.onload = () => resolve(img)
    img.src = src
  })
}

function makeTexture(canvas, maxAniso) {
  const tex = new THREE.CanvasTexture(canvas)
  tex.colorSpace = THREE.SRGBColorSpace
  tex.anisotropy = maxAniso
  return tex
}

// 纸杯腰封贴图：品牌蓝打底，正反两面各印一个横版 logo。
// 画布宽高比须等于腰封展开面的宽高比（周长 2π×0.74 : 高 0.75 ≈ 6.2:1），
// 否则贴图被横向拉伸、logo 显扁。
function makeBandCanvas(img) {
  const canvas = document.createElement('canvas')
  canvas.width = 2048
  canvas.height = 330
  const ctx = canvas.getContext('2d')
  ctx.fillStyle = BRAND_HEX
  ctx.fillRect(0, 0, canvas.width, canvas.height)
  const h = 280
  const w = h * (800 / 600)
  for (const cx of [canvas.width * 0.25, canvas.width * 0.75]) {
    ctx.drawImage(img, cx - w / 2, (canvas.height - h) / 2, w, h)
  }
  return canvas
}

// 透明杯贴图：透明底，只在 U=0.75 处印一个白鹿 logo
// （内层 rotation.y = π/2 后该位置朝向相机；只印一面，避免透过杯壁看到背面镜像）。
// 展开面宽高比：周长 2π×0.69 : 高 1.1 ≈ 3.94:1。
function makeDeerCanvas(img) {
  const canvas = document.createElement('canvas')
  canvas.width = 2048
  canvas.height = 520
  const ctx = canvas.getContext('2d')
  const h = 430
  const w = h * (1238 / 1542)
  ctx.drawImage(img, canvas.width * 0.75 - w / 2, (canvas.height - h) / 2, w, h)
  return canvas
}

// 纸杯：白杯身 + 品牌蓝腰封（logo 贴图）+ 蓝盖
function buildPaperCup(maxAniso) {
  const white = new THREE.MeshStandardMaterial({ color: 0xf7f8fc, roughness: 0.35 })
  const blue = new THREE.MeshStandardMaterial({ color: BRAND, roughness: 0.45 })
  const bandSide = new THREE.MeshStandardMaterial({ color: BRAND, roughness: 0.45 })

  // 杯身旋转成型，底部按真实纸杯结构：底盘内凹（抬高 0.08），
  // 侧壁纸延伸到底形成一圈贴地的窄裙边。
  // 剖面走向沿用「底面中心 → 外缘 → 顶面中心」保证法线朝外。
  const bodyProfile = [
    [0, -0.97], // 内凹的杯底中心
    [0.61, -0.97], // 凹底圆盘
    [0.61, -1.04], // 裙边内壁（竖直）
    [0.616, -1.05], // 折边小圆角
    [0.624, -1.05], // 裙边着地圈（边缘一圈纸）
    [0.643, -0.88], // 裙边段外壁：双层纸略外凸 0.004
    [0.639, -0.875], // 接缝台阶：内收回锥面，光照下显出一道淡线
    [0.85, 1.05], // 锥形外壁
    [0, 1.05], // 顶面（被杯盖遮挡）
  ].map(([x, y]) => new THREE.Vector2(x, y))
  const body = new THREE.Mesh(new THREE.LatheGeometry(bodyProfile, 48), white)
  const band = new THREE.Mesh(new THREE.CylinderGeometry(0.78, 0.7, 0.75, 48), [bandSide, blue, blue])
  band.position.y = -0.15
  const lidBase = new THREE.Mesh(new THREE.CylinderGeometry(0.9, 0.9, 0.16, 48), blue)
  lidBase.position.y = 1.13
  const lidTop = new THREE.Mesh(new THREE.CylinderGeometry(0.55, 0.82, 0.24, 48), blue)
  lidTop.position.y = 1.32
  const sip = new THREE.Mesh(new THREE.CylinderGeometry(0.18, 0.18, 0.1, 32), blue)
  sip.position.set(0, 1.48, 0.22)

  const inner = new THREE.Group()
  inner.add(body, band, lidBase, lidTop, sip)
  inner.rotation.y = Math.PI / 2 // 把腰封 logo 转到正面

  loadImage(logoBandUrl).then((img) => {
    bandSide.map = makeTexture(makeBandCanvas(img), maxAniso)
    bandSide.color.set(0xffffff) // 颜色与贴图相乘，置白以原色显示贴图
    bandSide.needsUpdate = true
  })

  const cup = new THREE.Group()
  cup.add(inner)
  return cup
}

// 透明冷饮杯：透明杯身印白鹿 logo + 透明平盖 + 居中吸管
function buildColdCup(maxAniso, env) {
  // 物理透射：真实的透明塑料质感（透光 + 折射），而非透明度混合的发白效果
  const clear = new THREE.MeshPhysicalMaterial({
    transmission: 1,
    transparent: true,
    roughness: 0.06,
    metalness: 0,
    ior: 1.45, // 塑料折射率
    thickness: 0.3, // 折射用壁厚
    attenuationColor: new THREE.Color(0xcdd8ff),
    attenuationDistance: 2.5,
    clearcoat: 0.5,
    envMap: env,
    envMapIntensity: 0.8,
    side: THREE.DoubleSide,
    depthWrite: false,
  })

  const body = new THREE.Mesh(new THREE.CylinderGeometry(0.8, 0.55, 2.0, 48), clear)
  body.renderOrder = 1

  // logo 印层：贴着杯身锥度的开口薄壳，透明贴图
  const shellMat = new THREE.MeshStandardMaterial({
    transparent: true,
    roughness: 0.4,
    depthWrite: false,
  })
  const shell = new THREE.Mesh(new THREE.CylinderGeometry(0.768, 0.631, 1.1, 48, 1, true), shellMat)
  shell.position.y = 0.1
  shell.renderOrder = 2

  // 平盖：旋转成型的透明薄盘，顶面约 1/2 半径处有一圈环形凹槽。
  // 凹槽截面为「宽 > 高」的扁矩形（参照真实瑞幸冷饮杯盖）：
  // 竖直槽壁 + 平槽底，仅在四个棱角处加一点圆角（用小段 45° 过渡近似）。
  // 注意走向：剖面须按「底面中心 → 外缘 → 顶面中心」排列，
  // LatheGeometry 的法线朝向由此决定，反了会内外翻转、反射全错。
  const lidProfile = [
    [0, -0.025],
    [0.527, -0.025], // 底面（槽圈以内）
    [0.527, -0.038], // 槽圈在盖底的对应下凸（薄壳结构，允许槽深超过盖厚）
    [0.633, -0.038],
    [0.633, -0.025],
    [0.85, -0.025], // 底面（槽圈以外）
    [0.85, 0.015], // 外缘
    [0.8, 0.025], // 外缘收边斜角
    [0.79, 0.025], // 凸圈外侧基部（凸圈紧贴最外缘）
    [0.79, 0.04], // 凸圈外壁（竖直）
    [0.787, 0.043], // 上棱圆角（对称小过渡）
    [0.743, 0.043], // 凸圈顶面（宽 0.05、高 0.018，矩形截面）
    [0.74, 0.04], // 上棱圆角（对称小过渡）
    [0.74, 0.025], // 凸圈内壁（竖直）
    [0.64, 0.025], // 竖纹平面（凸圈与凹槽之间）
    [0.625, 0.018], // 上棱圆角
    [0.625, -0.027], // 竖直槽壁
    [0.612, -0.03], // 下棱圆角
    [0.548, -0.03], // 平槽底（宽 0.09、深 0.055）
    [0.535, -0.027], // 下棱圆角
    [0.535, 0.018], // 竖直槽壁
    [0.52, 0.025], // 上棱圆角
    [0, 0.025], // 槽内平面到中心
  ].map(([x, y]) => new THREE.Vector2(x, y))
  // 竖纹：不堆几何面数，用程序生成的 bump 贴图扰动法线。
  // LatheGeometry 的 uv.y 按剖面点序号均分，竖纹只覆盖
  // 凸圈内壁与凹槽之间的环形平面（剖面点 13→14），
  // 盖面从外到内：凸圈 → 竖纹平面 → 凹槽 → 中心圆。
  const ribsCanvas = document.createElement('canvas')
  ribsCanvas.width = 1024
  ribsCanvas.height = 128
  {
    const ctx = ribsCanvas.getContext('2d')
    ctx.fillStyle = '#808080'
    ctx.fillRect(0, 0, ribsCanvas.width, ribsCanvas.height)
    const ribs = 120 // 偶数根，保证绕一圈首尾无缝
    const barW = ribsCanvas.width / ribs
    const segs = lidProfile.length - 1
    const y0 = Math.round((13 / segs) * ribsCanvas.height)
    const y1 = Math.round((14 / segs) * ribsCanvas.height)
    for (let i = 0; i < ribs; i++) {
      ctx.fillStyle = i % 2 ? '#9a9a9a' : '#6a6a6a'
      ctx.fillRect(Math.round(i * barW), y0, Math.ceil(barW), y1 - y0)
    }
  }
  const ribsTex = new THREE.CanvasTexture(ribsCanvas)
  ribsTex.flipY = false // 让 v 从剖面起点（底面中心）直接对应画布顶部
  ribsTex.wrapS = THREE.RepeatWrapping

  const lidMat = clear.clone()
  lidMat.bumpMap = ribsTex
  lidMat.bumpScale = 0.5
  // 盖子实际厚度仅 0.05，沿用杯身的 0.3 会导致透过盖子看东西时折射位移过度
  lidMat.thickness = 0.05
  // 盖子是法线朝外的封闭实体，只画正面：双面 + 关深度写入会产生正反面
  // 随机叠加的混合脏斑，反射看起来发花
  lidMat.side = THREE.FrontSide

  const lid = new THREE.Mesh(new THREE.LatheGeometry(lidProfile, 48), lidMat)
  lid.position.y = 1.04
  lid.renderOrder = 3

  // 吸管：穿过平盖中央，杯内一段透过杯壁可见
  // 下端伸到接近杯底（杯底 y=-1.0），上端保持伸出平盖
  const straw = new THREE.Mesh(
    new THREE.CylinderGeometry(0.06, 0.06, 2.9, 24),
    new THREE.MeshStandardMaterial({ color: 0xf7f8fc, roughness: 0.3 })
  )
  straw.position.y = 0.55
  straw.rotation.z = 0.1

  // ---- 美式咖啡液 ----
  // 渲染约束：咖啡必须是不透明材质——透射缓冲只画不透明物体，
  // transparent / transmission 的咖啡隔着透射杯壁都会消失或被杯身背面盖掉
  // （嵌套透射是引擎边缘路径，实测不可靠）。
  // 冰美式的「黑里透红」用菲涅尔边缘发光伪造：视角越掠射（液体边缘/薄层）
  // 越透出深红琥珀色，正面厚处保持近黑。
  // 半透明：普通透明混合。渲染顺序为 不透明 → 透射 → 透明，
  // 咖啡在透射杯壁之后绘制且杯壁不写深度，因此能平滑叠加在杯壁之上；
  // 吸管（不透明，先画）会按 1-opacity 透出来。
  // 注意不能用 transmission（与杯壁同队列互相覆盖）和 alphaHash（噪点）。
  const coffeeMat = new THREE.MeshPhysicalMaterial({
    color: 0x140a05, // 近黑的深咖啡
    roughness: 0.18,
    clearcoat: 0.9,
    clearcoatRoughness: 0.12,
    transparent: true,
    opacity: 0.8,
    depthWrite: false,
  })
  coffeeMat.onBeforeCompile = (shader) => {
    shader.fragmentShader = shader.fragmentShader.replace(
      '#include <emissivemap_fragment>',
      [
        '#include <emissivemap_fragment>',
        'float coffeeFres = pow(1.0 - saturate(dot(normalize(vViewPosition), normal)), 2.5);',
        'totalEmissiveRadiance += vec3(0.30, 0.05, 0.015) * coffeeFres;',
      ].join('\n')
    )
  }
  // 杯内壁（含 0.03 内缩防穿模）：r(y) = 0.55 + 0.125*(y+1) - 0.03，y 为杯局部坐标。
  // 液体从杯底 -0.97 灌到基准液面 0.45（约八分满）。
  const LIQ_BASE = -0.26 // 液体网格中心在杯局部系的高度
  const liquidGeo = new THREE.CylinderGeometry(0.701, 0.524, 1.42, 48)
  const liquid = new THREE.Mesh(liquidGeo, coffeeMat)
  liquid.position.y = LIQ_BASE
  const insetR = (cupY) => 0.55 + 0.125 * (cupY + 1) - 0.03

  // 顶面顶点索引（侧壁顶圈 + 顶盖），顶盖顶点另记下标用于更新法线
  const posAttr = liquidGeo.attributes.position
  const nrmAttr = liquidGeo.attributes.normal
  const basePos = posAttr.array.slice()
  const topIdx = []
  const capIdx = []
  for (let i = 0; i < posAttr.count; i++) {
    if (basePos[i * 3 + 1] > 1.42 / 2 - 1e-4) {
      topIdx.push(i)
      if (nrmAttr.array[i * 3 + 1] > 0.9) capIdx.push(i)
    }
  }

  // 液面晃动：小幅运动下液面始终近似平面 y = h + a·x + b·z，
  // 斜率 (a,b) 用带阻尼的弹簧振子追赶"视觉重力水平"——杯子动，液面滞后、过冲、回稳。
  const slosh = { a: 0, va: 0, b: 0, vb: 0 }
  const K = 40 // 刚度（晃动频率）
  const C = 4 // 阻尼（回稳快慢，欠阻尼带过冲）
  // 视觉重力：向下偏"远离观察者"，液面因此朝镜头展开，用户能看到液面
  const LIQUID_UP = new THREE.Vector3(0, 1, 0.45).normalize()
  const worldQ = new THREE.Quaternion()
  const up = new THREE.Vector3()

  const update = (dt) => {
    // 视觉重力的反方向在液体局部系中的方向 → 目标液面斜率
    liquid.getWorldQuaternion(worldQ)
    up.copy(LIQUID_UP).applyQuaternion(worldQ.invert())
    const ta = THREE.MathUtils.clamp(-up.x / up.y, -0.75, 0.75)
    const tb = THREE.MathUtils.clamp(-up.z / up.y, -0.75, 0.75)
    slosh.va += (-K * (slosh.a - ta) - C * slosh.va) * dt
    slosh.vb += (-K * (slosh.b - tb) - C * slosh.vb) * dt
    slosh.a = THREE.MathUtils.clamp(slosh.a + slosh.va * dt, -0.75, 0.75)
    slosh.b = THREE.MathUtils.clamp(slosh.b + slosh.vb * dt, -0.75, 0.75)

    for (const i of topIdx) {
      const x0 = basePos[i * 3]
      const y0 = basePos[i * 3 + 1]
      const z0 = basePos[i * 3 + 2]
      // 先按原位置算抬升，再把顶点半径缩放到该高度的锥形内壁上
      // （杯壁是锥面，不缩放会高侧露缝、低侧穿壁），最后用新半径修正高度
      const y1 = y0 + slosh.a * x0 + slosh.b * z0
      const s = insetR(y1 + LIQ_BASE) / insetR(y0 + LIQ_BASE)
      const x1 = x0 * s
      const z1 = z0 * s
      posAttr.array[i * 3] = x1
      posAttr.array[i * 3 + 1] = y0 + slosh.a * x1 + slosh.b * z1
      posAttr.array[i * 3 + 2] = z1
    }
    posAttr.needsUpdate = true
    // 顶盖法线统一为液面平面的法线（侧壁锥面法线只依赖角度，无需更新）
    const inv = 1 / Math.hypot(slosh.a, 1, slosh.b)
    for (const i of capIdx) {
      nrmAttr.array[i * 3] = -slosh.a * inv
      nrmAttr.array[i * 3 + 1] = inv
      nrmAttr.array[i * 3 + 2] = -slosh.b * inv
    }
    nrmAttr.needsUpdate = true
  }

  const inner = new THREE.Group()
  inner.add(body, shell, lid, straw, liquid)
  inner.rotation.y = Math.PI / 2 // 把杯身 logo 转到正面

  loadImage(logoDeerUrl).then((img) => {
    shellMat.map = makeTexture(makeDeerCanvas(img), maxAniso)
    shellMat.needsUpdate = true
  })

  const cup = new THREE.Group()
  cup.add(inner)
  return { cup, update }
}

onMounted(() => {
  const el = container.value
  scene = new THREE.Scene()
  camera = new THREE.PerspectiveCamera(35, el.clientWidth / el.clientHeight, 0.1, 100)
  // 相机距离自适应：垂直视场角固定，窄屏下水平视野变窄，
  // 双杯（含倾斜，半宽约 2.35 世界单位）会被画布边缘截断，
  // 此时后撤相机直到横向装得下；宽屏保持基准距离 6.2。
  const fitCamera = () => {
    const aspect = el.clientWidth / el.clientHeight
    camera.aspect = aspect
    const halfTan = Math.tan(THREE.MathUtils.degToRad(35 / 2)) * aspect
    camera.position.set(0, 0.9, Math.max(6.2, 2.35 / halfTan + 0.5))
    camera.lookAt(0, 0.15, 0)
    camera.updateProjectionMatrix()
  }
  fitCamera()

  renderer = new THREE.WebGLRenderer({ antialias: true, alpha: true })
  renderer.setSize(el.clientWidth, el.clientHeight)
  renderer.setPixelRatio(Math.min(window.devicePixelRatio, 2))
  el.appendChild(renderer.domElement)

  // 环境贴图只挂在透明材质上（不进 scene.environment，避免影响纸杯观感）
  pmrem = new THREE.PMREMGenerator(renderer)
  envTex = pmrem.fromScene(new RoomEnvironment(), 0.04).texture

  // 场景背景 = 纯品牌蓝，与 hero 在画布区域的 CSS 纯色严格一致（恒等无缝），
  // 同时给物理透射提供采样内容（CSS 背景透不进 WebGL）
  scene.background = new THREE.Color(BRAND)

  scene.add(new THREE.AmbientLight(0xffffff, 0.75))
  const key = new THREE.DirectionalLight(0xffffff, 1.6)
  key.position.set(3, 5, 4)
  scene.add(key)
  const rim = new THREE.DirectionalLight(0x9db4ff, 0.9)
  rim.position.set(-4, 2, -3)
  scene.add(rim)

  const maxAniso = renderer.capabilities.getMaxAnisotropy()

  // 纸杯居左：杯顶倾向右上方、远离观察者
  const paperCup = buildPaperCup(maxAniso)
  paperCup.position.set(-0.8, 0, -0.5)
  paperCup.rotation.set(-0.35, 0, -0.42)

  // 透明杯居右：杯顶倾向左上方、接近观察者
  const { cup: coldCup, update: updateCoffee } = buildColdCup(maxAniso, envTex)
  coldCup.position.set(0.8, 0, 0.5)
  coldCup.rotation.set(0.35, 0, 0.42)

  // 舞台层：鼠标视差作用于整体
  stage = new THREE.Group()
  stage.add(paperCup, coldCup)
  scene.add(stage)

  const mouse = { x: 0, y: 0 }
  onMouseMove = (e) => {
    mouse.x = (e.clientX / window.innerWidth) * 2 - 1
    mouse.y = (e.clientY / window.innerHeight) * 2 - 1
  }
  window.addEventListener('mousemove', onMouseMove)

  const clock = new THREE.Clock()
  const tick = () => {
    const dt = Math.min(clock.getDelta(), 0.05) // 切后台回来时夹住大步长，防弹簧爆掉
    const targetY = mouse.x * 0.18
    const targetX = mouse.y * 0.1
    stage.rotation.y += (targetY - stage.rotation.y) * 0.06
    stage.rotation.x += (targetX - stage.rotation.x) * 0.06
    updateCoffee(dt)
    renderer.render(scene, camera)
    raf = requestAnimationFrame(tick)
  }
  tick()

  onResize = () => {
    fitCamera()
    renderer.setSize(el.clientWidth, el.clientHeight)
  }
  window.addEventListener('resize', onResize)
})

onBeforeUnmount(() => {
  cancelAnimationFrame(raf)
  window.removeEventListener('resize', onResize)
  window.removeEventListener('mousemove', onMouseMove)
  envTex?.dispose()
  pmrem?.dispose()
  scene?.traverse((o) => {
    o.geometry?.dispose()
    const mats = Array.isArray(o.material) ? o.material : o.material ? [o.material] : []
    for (const m of mats) {
      m.map?.dispose()
      m.dispose()
    }
  })
  renderer?.dispose()
})
</script>

<template>
  <div ref="container" class="cup3d"></div>
</template>

<style scoped>
.cup3d {
  width: 100%;
  height: 450px; /* 高度即缩放：垂直视场角固定，画布越高杯子像素尺寸越大 */
}
</style>
