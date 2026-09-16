<template>
  <section class="est" aria-hidden="true">
    <div>
      <span class="sk lbl" />
      <span class="sk p50" />
    </div>
    <div class="track">
      <span class="sk tk lo" />
      <span class="sk tk hi" />
    </div>
    <span class="sk why" />
  </section>
</template>

<style scoped>
.est {
  background: var(--surface); border: 1px solid var(--border); border-radius: var(--r-lg);
  padding: 22px 24px; box-shadow: var(--sh-2); display: grid; gap: 14px;
}
.lbl { display: block; width: 188px; height: 14px; }
.p50 { display: block; width: 216px; height: 40px; margin-top: 8px; }
.track { position: relative; height: 10px; background: var(--surface-sunk); border-radius: 5px; margin: 22px 0 32px; }
.tk { position: absolute; top: 18px; height: 13px; width: 68px; }
.tk.lo { left: 0; }
.tk.hi { right: 0; }
.why { width: 194px; height: 15px; }
</style>
