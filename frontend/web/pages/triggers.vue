<script setup lang="ts">
useHead({ title: 'Триггеры — Номерограф' })
</script>

<template>
  <ComingSoon
    title="Следить за номером"
    lead="Подписка на появление номера в продаже."
  >
    <section>
      <h2>Зачем он вообще нужен</h2>
      <p>
        Шанс, что искомый номер продаётся прямо сейчас, — примерно один к семистам.
        Поэтому «ничего не найдено» — это
        не редкая неудача, а обычный исход поиска, и подписка на появление
        оказывается для нас не второстепенной функцией, а главным ответом.
      </p>
    </section>

    <section>
      <h2>Что он будет уметь</h2>
      <ul class="list">
        <li>Следить за конкретным номером — сообщим, как только он появится у любой из площадок.</li>
        <li>Следить за маской вида <code>А*7*AA77</code> и подходящим комбинациями.</li>
        <li>Ограничивать цену сверху, чтобы не приходили заведомо дорогие варианты.</li>
      </ul>
    </section>

    <template #exits>
      <NuxtLink to="/estimate" class="btn-primary">Пока оценить номер</NuxtLink>
      <NuxtLink to="/" class="btn-quiet">Посмотреть, что в продаже</NuxtLink>
    </template>
  </ComingSoon>
</template>

<style scoped>
section { display: grid; gap: 10px; }
h2 { font-size: 20px; }
p { color: var(--text-muted); }
.list { margin: 0; padding-left: 22px; display: grid; gap: 8px; color: var(--text-muted); }
code { font-family: var(--font-plate); font-weight: 700; background: var(--surface-sunk); padding: 1px 5px; border-radius: 2px; }

.btn-primary, .btn-quiet {
  min-height: 44px; padding: 11px 20px; display: inline-flex; align-items: center;
  border-radius: var(--r-md); font-size: 16px; font-weight: 600; text-decoration: none;
}
.btn-primary { background: var(--accent); color: #fff; }
.btn-primary:hover { background: var(--accent-hover); color: #fff; text-decoration: none; }
.btn-quiet { background: var(--surface); color: var(--text); border: 1px solid var(--border-strong); }
.btn-quiet:hover { border-color: var(--text-muted); color: var(--text); text-decoration: none; }
</style>
