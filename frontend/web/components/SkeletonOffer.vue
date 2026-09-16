<template>
  <div class="offer" aria-hidden="true">
    <div class="top">
      <span class="sk price" />
      <span class="sk src" />
    </div>
    <span class="sk tag" />
    <div class="rows">
      <span class="sk dt" /><span class="sk dd" />
      <span class="sk dt" /><span class="sk dd" />
    </div>
    <span class="sk go" />
  </div>
</template>

<style scoped>
.offer {
  background: var(--surface); border: 1px solid var(--border); border-left: 3px solid var(--border-strong);
  border-radius: var(--r-md); padding: 16px 18px; box-shadow: var(--sh-1);
  display: grid; gap: 11px; justify-items: start;
}
.top { display: flex; justify-content: space-between; align-items: baseline; gap: 14px; width: 100%; }
.rows { display: grid; grid-template-columns: auto 1fr; gap: 5px 14px; width: 100%; }
.price { width: 128px; height: 23px; }
.src   { width: 104px; height: 14px; }
.tag   { width: 138px; height: 20px; border-radius: var(--r-sm); }
.dt    { width: 92px;  height: 15px; }
.dd    { width: 124px; height: 15px; }
.go    { width: 158px; height: 15px; }
</style>
