<template>
  <div class="hold" aria-hidden="true">
    <div class="row">
      <span class="sk sign" />
      <span class="meta">
        <span class="sk region" />
        <span class="sk tag" />
        <span class="sk sub" />
      </span>
      <span class="right">
        <span class="sk price" />
        <span class="sk go" />
      </span>
    </div>
  </div>
</template>

<style scoped>
.hold { container-type: inline-size; display: grid; }

.row {
  background: var(--surface); border: 1px solid var(--border); border-radius: var(--r-lg);
  padding: 20px 24px; box-shadow: var(--sh-1);
  display: grid; grid-template-columns: auto 1fr auto; grid-template-areas: 'sign meta price';
  align-items: center; gap: 12px 28px;
}
.meta { grid-area: meta; display: grid; gap: 8px; justify-items: start; }
.right { grid-area: price; display: grid; gap: 8px; justify-items: end; }

.sign   { grid-area: sign; width: 218px; height: 62px; border-radius: 6px; }
.region { width: 92px;  height: 18px; }
.tag    { width: 138px; height: 25px; }
.sub    { width: 176px; height: 16px; }
.price  { width: 128px; height: 26px; }
.go     { width: 104px; height: 15px; }

@container (max-width: 760px) {
  .row {
    grid-template-columns: 1fr auto; grid-template-areas: 'sign sign' 'meta price';
    padding: 16px 18px; gap: 16px; align-items: end;
  }
}
@container (max-width: 440px) {
  .row { grid-template-columns: 1fr; grid-template-areas: 'sign' 'meta' 'price'; gap: 14px; }
  .right { justify-items: start; }
}

/* Знак мельчает по ширине экрана, а не колонки — так же, как настоящий PlateSign */
@media (max-width: 620px) {
  .sign { width: 172px; height: 50px; }
}
</style>
