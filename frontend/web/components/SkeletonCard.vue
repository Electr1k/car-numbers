<template>
  <div class="card" aria-hidden="true">
    <div class="head">
      <span class="sk plate" />
      <span class="sk price" />
    </div>
    <div class="sub">
      <span class="sk tag" />
      <span class="sk region" />
    </div>
    <div class="foot">
      <span class="sk meta" />
      <span class="sk go" />
    </div>
  </div>
</template>

<style scoped>
.card {
  background: var(--surface); border: 1px solid var(--border); border-radius: var(--r-lg);
  padding: 18px 20px; box-shadow: var(--sh-1);
  display: grid; gap: 12px; align-content: start;
}
.head { display: flex; justify-content: space-between; align-items: flex-start; gap: 16px; }
.sub { display: flex; gap: 14px; align-items: center; }
.foot {
  display: flex; justify-content: space-between; align-items: center; gap: 14px;
  border-top: 1px solid var(--border); padding-top: 12px;
}
.plate  { width: 148px; height: 38px; border-radius: var(--r-md); }
.price  { width: 104px; height: 26px; }
.tag    { width: 122px; height: 17px; }
.region { width: 88px;  height: 17px; }
.meta   { width: 168px; height: 15px; }
.go     { width: 132px; height: 15px; }
</style>
