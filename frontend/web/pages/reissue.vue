<script setup lang="ts">
/**
 * Самая содержательная страница продукта: о переоформлении пишет четверть
 * всех продавцов, и объяснить процедуру лучше конкурентов — реальное преимущество.
 * Источник фактов — docs/frontend/audience.md, §2.
 */
const steps = [
  {
    who: 'Покупатель',
    text: 'Покупает у продавца автомобиль вместе с нужным номером по договору купли-продажи.'
  },
  {
    who: 'Покупатель',
    text: 'В МРЭО меняет знаки на новые, а красивый номер оставляет «на хранение» — теперь он закреплён за покупателем.'
  },
  {
    who: 'Обе стороны',
    text: 'Тот же автомобиль продаётся обратно продавцу по второму договору, продавец ставит его на свой учёт.'
  },
  {
    who: 'Покупатель',
    text: 'Регистрирует собственный автомобиль и получает на него сохранённый номер.'
  }
]

const glossary = [
  {
    term: 'Перевес',
    means: 'Вся процедура целиком — та, что описана выше. Продавцы называют её так почти всегда.'
  },
  {
    term: 'Через моё авто',
    means: 'Продавец предоставляет свой автомобиль как посредника в сделке. Встречается в каждом десятом объявлении.'
  },
  {
    term: 'Без лишних записей в ПТС',
    means: 'Схема с третьим автомобилем: в паспорте вашей машины не останется отметок о промежуточных владельцах.'
  },
  {
    term: 'Переоформление включено',
    means: 'Продавец берёт процедуру на себя, и её стоимость уже в цене. У разных площадок это устроено по-разному — смотрите метку у цены.'
  },
  {
    term: 'На хранении',
    means: 'Номер уже снят с автомобиля и закреплён за владельцем в ГИБДД. Такую сделку провести проще.'
  }
]

useHead({
  title: 'Переоформление номера — Номерограф',
  meta: [{
    name: 'description',
    content: 'Как госномер юридически переходит к новому владельцу: четыре шага, два договора, ограничение по региону.'
  }]
})
</script>

<template>
  <div class="container page">
    <header class="head">
      <h1>Как номер переходит к новому владельцу</h1>
      <p class="lede">
        Продать регистрационный знак отдельно от автомобиля закон не позволяет.
        Рынок работает по обходной схеме, и она сложнее, чем кажется со стороны:
        четыре шага, два договора и два визита в МРЭО.
      </p>
    </header>

    <section class="block">
      <h2>Четыре шага</h2>
      <ol class="steps">
        <li v-for="(s, i) in steps" :key="i">
          <span class="n" aria-hidden="true">{{ i + 1 }}</span>
          <div>
            <span class="who">{{ s.who }}</span>
            <p>{{ s.text }}</p>
          </div>
        </li>
      </ol>
    </section>

    <aside class="warn">
      <h2>Главное ограничение</h2>
      <p>
        С 2020 года знак выдаётся с кодом региона по месту регистрации владельца,
        и обе стороны сделки должны быть прописаны в одном регионе.
        Поэтому номер с кодом 761 доступен жителю Ростовской области —
        и недоступен москвичу, как бы он ни понравился.
      </p>
      <p class="warn-note">
        Именно поэтому код региона виден на каждой карточке: это не украшение, а предупреждение.
      </p>
    </aside>

    <section class="block">
      <h2>Вариант без лишних записей в ПТС</h2>
      <p class="text">
        Существует схема с третьим автомобилем: он выступает промежуточным звеном,
        и в паспорте вашей машины не остаётся отметок о лишних владельцах.
        Это дороже и дольше, но некоторых покупателей смущает именно история в документах.
        Продавцы указывают такую возможность прямо в объявлении.
      </p>
    </section>

    <section class="block">
      <h2>Словарь объявлений</h2>
      <p class="text">
        Продавцы пишут на своём языке, и человеку, впервые попавшему на этот рынок,
        объявления читаются с трудом. Вот что стоит за частыми формулировками.
      </p>
      <dl class="glossary">
        <div v-for="g in glossary" :key="g.term">
          <dt>{{ g.term }}</dt>
          <dd>{{ g.means }}</dd>
        </div>
      </dl>
    </section>

    <section class="block">
      <h2>Что будет, когда номера начнут продавать официально</h2>
      <p class="text">
        Законопроект о продаже государственных знаков через «Госуслуги» внесён в Госдуму
        и поддержан на высшем уровне, но <strong>пока не принят</strong>. Обсуждается
        продажа по установленной цене и через аукцион. До вступления поправок в силу
        перевес остаётся единственным работающим способом получить конкретный номер.
      </p>
    </section>

    <section class="cta">
      <h2>Что дальше</h2>
      <div class="exits">
        <NuxtLink to="/" class="btn-primary">Подобрать номер</NuxtLink>
        <NuxtLink to="/estimate" class="btn-quiet">Оценить свой номер</NuxtLink>
      </div>
    </section>

    <p class="disclaimer">
      Страница объясняет сложившуюся практику рынка и не является юридической консультацией.
      Порядок регистрации транспортных средств устанавливает МВД, и он может измениться.
    </p>
  </div>
</template>

<style scoped>
.page { padding: 52px 24px 64px; max-width: 820px; }

.head { margin-bottom: 40px; }
.lede { font-size: 18px; color: var(--text-muted); margin-top: 14px; }

.block { margin-bottom: 40px; }
.block h2 { margin-bottom: 16px; }
.text { color: var(--text-muted); }

/* Нумерация здесь честная: это настоящая последовательность действий */
.steps { list-style: none; margin: 0; padding: 0; counter-reset: step; }
.steps li {
  display: grid; grid-template-columns: 44px 1fr; gap: 18px; align-items: start;
  padding: 18px 0; border-bottom: 1px solid var(--border);
}
.steps li:first-child { padding-top: 0; }
.steps li:last-child { border-bottom: none; padding-bottom: 0; }
.n {
  font-family: var(--font-plate); font-weight: 700; font-size: 28px; line-height: 1.1;
  color: var(--accent); text-align: right;
}
.who {
  display: block; font-size: 12px; font-weight: 700; text-transform: uppercase;
  letter-spacing: .1em; color: var(--text-faint); margin-bottom: 4px;
}
.steps p { color: var(--text); }

.warn {
  background: var(--warn-bg); border-radius: var(--r-lg); padding: 24px 26px;
  margin-bottom: 40px; display: grid; gap: 12px;
}
.warn h2 { color: var(--warn); font-size: 20px; margin: 0; }
.warn p { color: var(--warn); }
.warn-note { font-size: 15px; opacity: .85; }

.glossary { margin: 0; display: grid; gap: 0; }
.glossary > div {
  display: grid; grid-template-columns: minmax(180px, 260px) 1fr; gap: 20px;
  padding: 16px 0; border-bottom: 1px solid var(--border);
}
.glossary > div:last-child { border-bottom: none; }
dt { font-weight: 600; }
dd { margin: 0; color: var(--text-muted); font-size: 15.5px; }

.cta { margin-bottom: 36px; }
.cta h2 { margin-bottom: 16px; }
.exits { display: flex; flex-wrap: wrap; gap: 12px; }
.btn-primary, .btn-quiet {
  min-height: 44px; padding: 11px 20px; display: inline-flex; align-items: center;
  border-radius: var(--r-md); font-size: 16px; font-weight: 600; text-decoration: none;
}
.btn-primary { background: var(--accent); color: #fff; }
.btn-primary:hover { background: var(--accent-hover); color: #fff; text-decoration: none; }
.btn-quiet { background: var(--surface); color: var(--text); border: 1px solid var(--border-strong); }
.btn-quiet:hover { border-color: var(--text-muted); color: var(--text); text-decoration: none; }

.disclaimer {
  font-size: 14.5px; color: var(--text-faint); line-height: 1.5;
  border-top: 1px solid var(--border); padding-top: 20px;
}

@media (max-width: 620px) {
  .glossary > div { grid-template-columns: 1fr; gap: 6px; }
  .steps li { grid-template-columns: 34px 1fr; gap: 14px; }
  .n { font-size: 24px; }
}
</style>
