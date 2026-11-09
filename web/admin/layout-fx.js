// Layout reactions (settings.layoutFx): switches on the Layout page for the picked
// layout. The value holds the ids of every layout, so switching layouts keeps them.
(() => {
  // [id, kind, ru name, ru hint, en name, en hint]; kind: win | loss | idle | detail.
  const DEFS = {
    detailed: [
      ['detailed-win', 'win', 'Прыжок', 'Панель подпрыгивает и мягко приземляется.', 'Hop', 'The panel hops and lands softly.'],
      ['detailed-loss', 'loss', 'Отдача', 'Панель отбрасывает назад, как от попадания.', 'Recoil', 'The panel recoils as if hit.'],
      ['detailed-idle', 'idle', 'Покачивание', 'Еле заметно поднимается и опускается.', 'Bob', 'Rises and sinks a little.'],
      ['det-trace', 'detail', 'Обводка по рамке', 'По контуру панели пробегает свет цвета результата.', 'Edge trace', 'A light in the result colour runs around the edge.'],
    ],
    mini: [
      ['mini-win', 'win', 'Резинка', 'Растягивается, как резинка, и пружинит обратно.', 'Rubber band', 'Stretches like a rubber band and springs back.'],
      ['mini-loss', 'loss', 'Сдувается', 'Сплющивается и медленно надувается.', 'Deflate', 'Squashes down and slowly fills up again.'],
      ['mini-idle', 'idle', 'Сердцебиение', 'Раз в пару секунд — два коротких удара.', 'Heartbeat', 'Two short beats every few seconds.'],
      ['mini-coin', 'detail', 'Монетка', 'Буква W или L делает два оборота, когда цифра досчитала.', 'Coin', 'The W or L spins twice once the number has counted.'],
      ['mini-grow', 'detail', 'Вдох', 'По краю изнутри вспыхивает свечение цвета результата.', 'Breath', 'A glow in the result colour flashes along the inner edge.'],
    ],
    card: [
      ['card-win', 'win', 'Кивок к зрителю', 'Карточка кивает верхом к зрителю.', 'Nod', 'The card tips its top towards the viewer.'],
      ['card-loss', 'loss', 'Отшатнулся', 'Карточка отшатывается назад.', 'Flinch', 'The card leans back and returns.'],
      ['card-idle', 'idle', 'Висит на гвозде', 'Покачивается, будто висит на гвоздике.', 'Hanging', 'Sways as if hanging on a nail.'],
      ['card-wr', 'detail', 'Бегущий винрейт', 'По полосе WR пробегает блик.', 'Win-rate glint', 'A glint runs along the win-rate bar.'],
      ['card-curtain', 'detail', 'Шторка', 'Сверху вниз проходит шторка цвета результата.', 'Curtain', 'A curtain in the result colour passes down.'],
    ],
    strip: [
      ['strip-win', 'win', 'Рывок ленты', 'Лента дёргается вперёд и пружинит на место.', 'Jolt', 'The strip jolts forward and springs back.'],
      ['strip-loss', 'loss', 'Помеха', 'Лента дрожит рывками, как сбитый сигнал.', 'Interference', 'The strip jitters like a bad signal.'],
      ['strip-live', 'detail', 'LIVE-точка', 'В углу мигает красная точка, как в эфире.', 'LIVE dot', 'A red on-air dot blinks in the corner.'],
    ],
    minimal: [
      ['minimal-win', 'win', 'Взлёт', 'Поднимается, зависает и опускается.', 'Lift-off', 'Rises, hangs there and comes down.'],
      ['minimal-loss', 'loss', 'Падение', 'Падает вниз и пару раз подпрыгивает.', 'Drop', 'Drops and bounces a couple of times.'],
      ['minimal-idle', 'idle', 'Парение', 'Плавно дрейфует над игрой.', 'Hover', 'Drifts gently over the game.'],
      ['min-dmg', 'detail', '+1 как урон', 'Сбоку от цифры всплывает «+1».', '+1 damage number', 'A "+1" floats up next to the number.'],
    ],
    column: [
      ['column-win', 'win', 'Вытягивается', 'Колонка вытягивается вверх.', 'Stretch', 'The column stretches up.'],
      ['column-loss', 'loss', 'Сжимается', 'Колонка сжимается вниз.', 'Squash', 'The column squashes down.'],
      ['column-idle', 'idle', 'Маятник', 'Покачивается, как маятник.', 'Pendulum', 'Sways like a pendulum.'],
      ['col-lift', 'detail', 'Лифт', 'Полоса света едет вверх при победе, вниз при поражении.', 'Lift', 'A band of light rides up on a win, down on a loss.'],
    ],
    ring: [
      ['ring-win', 'win', 'Мячик', 'Подпрыгивает дважды, как мячик.', 'Ball', 'Bounces twice like a ball.'],
      ['ring-loss', 'loss', 'Откат', 'Откатывается назад, как колесо.', 'Roll back', 'Rolls back like a wheel and returns.'],
      ['ring-idle', 'idle', 'Покачивание колеса', 'Чуть покачивается на месте.', 'Wheel rock', 'Rocks a little in place.'],
      ['ring-breath', 'detail', 'Дыхание серии', 'Ореол вокруг кольца: чем длиннее серия побед, тем ярче.', 'Streak breath', 'A halo around the ring, brighter with a longer win streak.'],
      ['ring-mark', 'detail', 'Метка последней игры', 'Метка на конце дуги переезжает к новому винрейту.', 'Last game mark', 'A mark at the end of the arc moves to the new win rate.'],
    ],
    slant: [
      ['slant-win', 'win', 'Рывок вперёд', 'Рывок вперёд, как рывок героя.', 'Dash', 'Dashes forward like a hero.'],
      ['slant-loss', 'loss', 'Оглушение', 'Покачивается, будто оглушён.', 'Stun', 'Wobbles as if stunned.'],
      ['slant-idle', 'idle', 'Наклон HUD', 'Стоит с наклоном вглубь и чуть «дышит».', 'HUD tilt', 'Stands tilted into the screen and breathes a little.'],
    ],
    scoreboard: [
      ['scoreboard-win', 'win', 'Рёв трибун', 'Табло вибрирует, как от рёва трибун.', 'Crowd roar', 'The board shakes like a roaring crowd.'],
      ['scoreboard-loss', 'loss', 'Просадка питания', 'Табло мигает, будто просело питание.', 'Power dip', 'The board flickers as if the power dipped.'],
      ['sb-led', 'detail', 'LED-табло', 'Поверх цифры мерцает сетка лампочек.', 'LED board', 'A grid of bulbs flickers over the number.'],
    ],
    session: [
      ['session-win', 'win', 'Шаг вверх', 'Поднимается на ступеньку и опускается.', 'Step up', 'Steps up, holds, and comes back.'],
      ['session-loss', 'loss', 'Шаг вниз', 'Опускается на ступеньку и поднимается.', 'Step down', 'Steps down, holds, and comes back.'],
      ['ses-label', 'detail', 'Подсветка SESSION', 'Подпись SESSION вспыхивает цветом результата.', 'SESSION highlight', 'The SESSION label flashes in the result colour.'],
    ],
    form: [
      ['form-win', 'win', 'Кивок «да»', 'Кивает «да».', 'Nod "yes"', 'Nods "yes".'],
      ['form-loss', 'loss', 'Мотает «нет»', 'Мотает «нет».', 'Shake "no"', 'Shakes "no".'],
      ['form-idle', 'idle', 'Ожидание', 'Чуть поворачивается из стороны в сторону.', 'Waiting', 'Turns a little from side to side.'],
      ['form-belt', 'detail', 'Конвейер', 'Старая плитка уходит, новая падает сверху. Вместе с «Домино» не работает.', 'Conveyor', 'The oldest tile leaves, the new one drops in. Not together with Domino.'],
      ['form-domino', 'detail', 'Домино', 'Плитки по очереди переворачиваются. Вместе с «Конвейером» не работает.', 'Domino', 'The tiles flip one after another. Not together with Conveyor.'],
    ],
    ladder: [
      ['ladder-win', 'win', 'Ступенька', 'Поднимается в два шага.', 'Climb', 'Climbs up in two steps.'],
      ['ladder-loss', 'loss', 'Оступился', 'Оступается вниз и выравнивается.', 'Slip', 'Slips down and recovers.'],
      ['lad-near', 'detail', 'Манит следующий тир', 'Когда до нового тира 1–2 шага, его иконка пульсирует.', 'Next tier calls', 'The next tier icon pulses when it is 1–2 steps away.'],
      ['lad-piano', 'detail', 'Клавиши', 'Ступени нажимаются по очереди, как клавиши.', 'Keys', 'The steps are pressed one by one like keys.'],
    ],
    roles: [
      ['roles-win', 'win', 'Выпад', 'Делает выпад вперёд.', 'Lunge', 'Lunges forward.'],
      ['roles-loss', 'loss', 'Вздрог', 'Вздрагивает.', 'Flinch', 'Flinches.'],
      ['rol-slide', 'detail', 'Переезд подсветки', 'При смене роли подсветка плавно переезжает.', 'Sliding highlight', 'The highlight slides to the new role.'],
      ['rol-scan', 'detail', 'Скан строки', 'После игры по активной строке проходит луч.', 'Row scan', 'A beam sweeps the played row after a game.'],
      ['rol-dim', 'detail', 'Фокус на роли', 'Другие роли серые и размытые.', 'Role focus', 'The other roles go grey and soft.'],
    ],
    hex: [
      ['hex-loss', 'loss', 'Удар по соте', 'Сжимается от удара и выпрямляется.', 'Hit', 'Squeezes from a hit and straightens.'],
      ['hex-idle', 'idle', 'Пульс соты', 'Едва заметно пульсирует.', 'Pulse', 'Pulses very slightly.'],
      ['hex-edges', 'detail', 'Грани серии', 'Каждая победа подряд зажигает одну грань.', 'Streak edges', 'Each win in a row lights one edge.'],
      ['hex-pulse', 'detail', 'Пульс контура', 'Контур вспыхивает цветом результата.', 'Contour flash', 'The outline flashes in the result colour.'],
    ],
    feed: [
      ['feed-win', 'win', 'Влёт', 'Влетает справа, как новая строка килфида.', 'Slide in', 'Slides in from the right like a new kill.'],
      ['feed-loss', 'loss', 'Тряска фида', 'Трясётся рывками.', 'Shake', 'Shakes in jolts.'],
      ['feed-flash', 'detail', 'Белая вспышка', 'Новая строка вспыхивает белым.', 'White flash', 'The new row flashes white.'],
      ['feed-cross', 'detail', 'Прицел серии', 'С трёх побед подряд появляется прицел и счётчик.', 'Streak crosshair', 'From three wins in a row a crosshair and count appear.'],
    ],
    tug: [
      ['tug-win', 'win', 'Тянет к победам', 'Канат тянет виджет в сторону побед.', 'Pulled to wins', 'The rope pulls the widget towards wins.'],
      ['tug-loss', 'loss', 'Тянет к поражениям', 'Канат тянет виджет в сторону поражений.', 'Pulled to losses', 'The rope pulls the widget towards losses.'],
      ['tug-jerk', 'detail', 'Рывок', 'Канат коротко дёргается.', 'Rope jerk', 'The rope gives a short jerk.'],
      ['tug-tension', 'detail', 'Натяжение', 'Натянутая часть каната наливается светом.', 'Tension', 'The pulled part of the rope fills with light.'],
    ],
    flip: [
      ['flip-win', 'win', 'Стук механизма', 'Табло стучит, как механизм с флапами.', 'Clack', 'The board clacks like a flap mechanism.'],
      ['flip-loss', 'loss', 'Стук механизма', 'Табло стучит, как механизм с флапами.', 'Clack', 'The board clacks like a flap mechanism.'],
      ['fl-word', 'detail', 'Табличка', 'Ряд флапов под табло показывает WIN или LOSS. Виджет становится выше.', 'Word row', 'A row of flaps under the board shows WIN or LOSS. Makes the widget taller.'],
    ],
  };
  const EXCLUSIVE = { 'form-belt': 'form-domino', 'form-domino': 'form-belt' };
  const KIND = {
    win: ['победа', 'win'], loss: ['поражение', 'loss'], idle: ['в покое', 'at rest'], detail: ['деталь', 'detail'],
  };
  let on = new Set();
  let filled = false;

  const lang = () => (window.AdminI18n?.getLang?.() === 'en' ? 1 : 0);
  const t = (key, ru, en) => {
    const v = window.AdminI18n?.t?.(key);
    return v && v !== key ? v : (lang() ? en : ru);
  };

  function fill(value) {
    on = new Set(String(value || '').split(',').map((s) => s.trim()).filter((s) => s && s !== 'none'));
    filled = true;
    const box = document.getElementById('layoutFxOpts');
    box?.querySelectorAll('input[name="layoutFx"]').forEach((b) => { b.checked = on.has(b.value); });
  }
  // Before the saved settings arrive, send nothing: the app fills in its defaults.
  function read() {
    if (!filled) return '';
    return on.size ? [...on].join(',') : 'none';
  }

  function opt([id, kind, ru, ruHint, en, enHint]) {
    const l = lang();
    return `<label class="card-opt"><input type="checkbox" name="layoutFx" value="${id}" data-kind="${kind}"${on.has(id) ? ' checked' : ''}>`
      + `<span><b>${l ? en : ru} <em class="lfx-kind lfx-${kind}">${KIND[kind][l]}</em></b><span>${l ? enHint : ruHint}</span></span></label>`;
  }

  function render(layout) {
    const box = document.getElementById('layoutFxOptions');
    const list = document.getElementById('layoutFxOpts');
    if (!box || !list) return;
    const defs = DEFS[layout];
    box.hidden = !defs;
    if (!defs) return;
    const motion = defs.filter((d) => d[1] !== 'detail');
    const details = defs.filter((d) => d[1] === 'detail');
    list.innerHTML = `<h3 class="lfx-group">${t('lfx.motion', 'Сам виджет', 'The widget itself')}</h3><div class="card-opts">${motion.map(opt).join('')}</div>`
      + (details.length ? `<h3 class="lfx-group">${t('lfx.details', 'Детали', 'Details')}</h3><div class="card-opts">${details.map(opt).join('')}</div>` : '');
  }

  // A switch changed: keep the value, enforce Conveyor / Domino, say which game to show.
  function onChange(input) {
    const id = input.value;
    if (input.checked) {
      on.add(id);
      const other = EXCLUSIVE[id];
      if (other) {
        on.delete(other);
        const box = document.querySelector(`input[name="layoutFx"][value="${other}"]`);
        if (box) box.checked = false;
      }
    } else {
      on.delete(id);
    }
    return input.dataset.kind;
  }

  window.AdminLayoutFx = { fill, read, render, onChange };
})();
