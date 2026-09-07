<script setup lang="ts">
import type { FeedResponse } from '~/types/api'

/* Примеры берём из ленты, чтобы ссылки вели на живые номера, а не на выдумку */
const { data: feed } = await useFetch<FeedResponse>('/api/v1/feed', { query: { limit: 3 } })

useHead({
  title: 'Оценка номера — Номерограф',
  meta: [{
    name: 'description',
    content: 'Сколько стоит любой госномер: интервал цены, разбор по признакам и объявления трёх площадок.'
  }]
})
</script>

<template>
  <div>
    <section class="hero container">
      <h1>Сколько стоит номер</h1>
      <p class="lede">
        Введите любой номер — даже тот, которого сейчас нет в продаже.
        Оценка строится по объявлениям трёх площадок и объясняет, из чего сложилась цифра.
      </p>
      <PlateSearch mode="estimate" label="Номер для оценки" cta="Оценить" />

      <p v-if="feed?.items?.length" class="examples">
        Например:
        <NuxtLink v-for="c in feed.items" :key="c.id" :to="`/valuation/${c.number}`">{{ c.number }}</NuxtLink>
      </p>
    </section>

    <section class="container band">
      <h2>Что вы получите</h2>
      <div class="cards">
        <article>
          <h3>Интервал, а не одно число</h3>
          <p>
            Рынок расходится в цене одного и того же номера. Одна цифра создавала бы ложное
            впечатление точности, поэтому мы показываем полосу, в которую попадает половина
            похожих объявлений.
          </p>
        </article>
        <article>
          <h3>Разбор до последнего множителя</h3>
          <p>
            Видно, что тянет цену вверх, а что вниз: регион, серия букв, узор цифр.
            Произведение множителей даёт итоговую цену — это настоящая раскладка, а не иллюстрация.
          </p>
        </article>
        <article>
          <h3>Цена продавца, а не витрины</h3>
          <p>
            Одна из площадок продаёт номер вместе с переоформлением и включает его в цену.
            Мы это вычитаем, чтобы объявления можно было сравнивать между собой.
          </p>
        </article>
      </div>
    </section>

    <section class="container band">
      <h2>Насколько ей можно верить</h2>
      <p class="sub">
        Цифры ниже — замеры на отложенной выборке, а не обещания.
        Проверка честная: свежие объявления от модели прячут, потом сверяют.
      </p>

      <dl class="stats">
        <div>
          <dt>Половина оценок ошибается меньше чем на</dt>
          <dd>21%</dd>
        </div>
        <div>
          <dt>Попадает в заявленный интервал</dt>
          <dd>54%</dd>
        </div>
        <div>
          <dt>Систематического перекоса</dt>
          <dd>нет</dd>
        </div>
      </dl>

      <h3 class="sub-h">Где мы слабее</h3>
      <table>
        <thead>
          <tr><th scope="col">Цена номера</th><th scope="col">Медианная ошибка</th></tr>
        </thead>
        <tbody>
          <tr><td>20 – 50 тыс. ₽</td><td class="good">17%</td></tr>
          <tr><td>50 – 500 тыс. ₽</td><td>23%</td></tr>
          <tr><td>0,5 – 2 млн ₽</td><td>26%</td></tr>
          <tr><td>дороже 2 млн ₽</td><td class="bad">41%</td></tr>
          <tr><td>дешевле 20 тыс. ₽</td><td class="bad">28%</td></tr>
        </tbody>
      </table>
      <p class="note">
        На дорогих и очень дешёвых номерах оценка помечается пониженным доверием прямо в ответе —
        мы не делаем вид, что знаем их так же хорошо.
      </p>
    </section>

    <section class="container band last">
      <h2>Чего оценка не знает</h2>
      <ul class="limits">
        <li>
          <strong>Это цена запроса, а не цена сделки.</strong> Мы учимся на том, за сколько
          номера выставляют, а не за сколько их покупают. Каждый восьмой продавец прямо пишет,
          что торг уместен.
        </li>
        <li>
          <strong>Номер можно поставить только на автомобиль своего региона.</strong>
          Оценка не зависит от того, где живёте вы, — но купить номер чужого региона не выйдет.
          <NuxtLink to="/reissue">Как устроено переоформление</NuxtLink>
        </li>
        <li>
          <strong>Неизвестные знаки расширяют интервал.</strong> Маску вида
          <code>А0*0АА761</code> мы раскрываем перебором, но при трёх и более пропусках
          отказываемся: разброс перестаёт нести смысл.
        </li>
      </ul>
    </section>
  </div>
</template>

<style scoped>
.hero { padding: 52px 24px 40px; }
.lede { font-size: 18px; color: var(--text-muted); margin: 14px 0 24px; max-width: 640px; }
.examples { margin-top: 20px; font-size: 15px; color: var(--text-faint); display: flex; flex-wrap: wrap; gap: 12px; align-items: baseline; }
.examples a { font-family: var(--font-plate); font-weight: 700; font-size: 18px; letter-spacing: .04em; }

.band { padding: 0 24px 44px; }
.band.last { padding-bottom: 64px; }
.band h2 { margin-bottom: 16px; }
.sub { color: var(--text-muted); max-width: 640px; margin-bottom: 20px; }
.sub-h { margin: 28px 0 12px; }

.cards { display: grid; grid-template-columns: repeat(auto-fit, minmax(280px, 1fr)); gap: 16px; }
.cards article {
  background: var(--surface); border: 1px solid var(--border); border-radius: var(--r-lg);
  padding: 20px 22px; box-shadow: var(--sh-1); display: grid; gap: 10px; align-content: start;
}
.cards p { color: var(--text-muted); font-size: 15.5px; }

.stats { display: grid; grid-template-columns: repeat(auto-fit, minmax(220px, 1fr)); gap: 16px; margin: 0 0 8px; }
.stats > div {
  background: var(--surface); border: 1px solid var(--border); border-radius: var(--r-lg);
  padding: 18px 20px; box-shadow: var(--sh-1);
}
.stats dt { font-size: 14.5px; color: var(--text-muted); }
.stats dd { margin: 6px 0 0; font-size: 34px; font-weight: 700; letter-spacing: -.03em; font-variant-numeric: tabular-nums; }

table { width: 100%; border-collapse: collapse; font-size: 15.5px; max-width: 560px; }
th {
  text-align: left; font-size: 12px; font-weight: 600; text-transform: uppercase; letter-spacing: .09em;
  color: var(--text-faint); padding: 0 14px 8px 0; border-bottom: 1px solid var(--border);
}
th:last-child, td:last-child { text-align: right; font-variant-numeric: tabular-nums; }
td { padding: 10px 14px 10px 0; border-bottom: 1px solid var(--border); }
tr:last-child td { border-bottom: none; }
.good { color: var(--good); font-weight: 700; }
.bad { color: var(--alert); font-weight: 700; }
.note { margin-top: 14px; font-size: 15px; color: var(--text-muted); max-width: 640px; }

.limits { list-style: none; margin: 0; padding: 0; display: grid; gap: 14px; max-width: 720px; }
.limits li {
  border-left: 3px solid var(--border-strong); padding-left: 16px;
  font-size: 15.5px; color: var(--text-muted); line-height: 1.55;
}
.limits strong { color: var(--text); }
.limits code { font-family: var(--font-plate); background: var(--surface-sunk); padding: 1px 5px; border-radius: 2px; }
</style>
