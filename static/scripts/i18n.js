// ============================================
// INTERNATIONALIZATION (i18n)
// ============================================

// Translation dictionary
const translations = {
    ru: {
        // Header
        'header.login': 'Начать',
        'header.myLists': 'Мои вишлисты',
        'header.profile': 'Профиль',
        'header.logout': 'Выйти',

        // Home page
        'home.title': 'Wishlist',
        'home.subtitle': 'Создавайте списки желаний и делитесь ими с близкими',
        'home.feature1.title': 'Организация',
        'home.feature1.desc': 'Создавайте списки для разных событий – дни рождения, праздники или мечты',
        'home.feature2.title': 'Контроль',
        'home.feature2.desc': 'Публичные и приватные списки с доступом только по ссылке',
        'home.feature3.title': 'Уведомления',
        'home.feature3.desc': 'Смотрите график цены и узнавайте о снижении',
        'home.feature3.badge': 'Однажды...',

        // Auth modal
        'auth.login': 'Вход',
        'auth.register': 'Регистрация',
        'auth.username': 'Юзернейм',
        'auth.password': 'Пароль',
        'auth.email': 'Почта',
        'auth.name': 'Имя',
        'auth.loginButton': 'Войти',
        'auth.registerButton': 'Создать аккаунт',
        'auth.forgotPassword': 'Забыли пароль?',
        'auth.forgotPasswordTitle': 'Восстановление пароля',
        'auth.forgotPasswordHint': 'Укажите email аккаунта, и мы отправим ссылку для сброса пароля',
        'auth.forgotPasswordSubmit': 'Отправить ссылку',
        'auth.forgotPasswordSuccess': 'Если аккаунт с таким email существует, письмо для сброса пароля уже отправлено',
        'auth.forgotPasswordError': 'Не удалось отправить письмо для сброса пароля',
        'auth.resetPasswordTitle': 'Новый пароль',
        'auth.resetPasswordHint': 'Введите новый пароль для вашего аккаунта',
        'auth.resetPasswordSubmit': 'Сохранить новый пароль',
        'auth.resetPasswordNew': 'Новый пароль',
        'auth.resetPasswordConfirm': 'Повторите пароль',
        'auth.resetPasswordSuccess': 'Пароль успешно изменён',
        'auth.resetPasswordError': 'Не удалось изменить пароль',
        'auth.resetPasswordMismatch': 'Пароли не совпадают',
        'auth.resetPasswordInvalidToken': 'Ссылка для сброса пароля недействительна или устарела',

        // Lists page
        'lists.title': 'Мои вишлисты',
        'lists.search': 'Поиск...',
        'lists.sort': 'Сортировка',
        'lists.filter': 'Фильтр',
        'lists.create': 'Создать список',
        'lists.empty': 'У вас пока нет списков',
        'lists.emptyTitle': 'У вас пока нет списков',
        'lists.emptySubtitle': 'Создайте свой первый список желаний',
        'lists.sort.newFirst': 'Сначала новые',
        'lists.sort.oldFirst': 'Сначала старые',
        'lists.sort.nameAsc': 'По имени (А-Я)',
        'lists.sort.nameDesc': 'По имени (Я-А)',
        'lists.sort.updatedDesc': 'По обновлению (новые)',
        'lists.sort.updatedAsc': 'По обновлению (старые)',
        'lists.sort.wishesDesc': 'По количеству желаний (убыв.)',
        'lists.sort.wishesAsc': 'По количеству желаний (возр.)',
        'lists.filter.all': 'Все списки',
        'lists.filter.public': 'Публичные',
        'lists.filter.private': 'Приватные',
        'lists.filter.empty': 'Пустые',
        'lists.filter.notEmpty': 'Непустые',
        'lists.loadFailed': 'Не удалось загрузить списки',
        'lists.back': 'Назад',
        'users.find.title': 'Найти пользователя',
        'users.find.placeholder': 'Введите юзернейм',
        'users.find.submit': 'Открыть публичные списки',
        'users.find.notFound': 'Пользователь не найден',
        'users.find.failed': 'Не удалось найти пользователя',
        'users.find.searchEmpty': 'Начните вводить юзернейм',
        'users.find.searchNoResults': 'Ничего не найдено',
        'users.find.searchChoose': 'Выберите пользователя из списка',
        'users.publicLists.back': 'К моим вишлистам',
        'users.publicLists.subtitle': 'Здесь показаны только публичные списки',
        'users.publicLists.emptyTitle': 'Тут пусто',
        'users.publicLists.emptySubtitle': 'Похоже, у этого пользователя ещё нет вишлистов',

        // List detail
        'list.copyLink': 'Скопировать ссылку',
        'list.generateLink': 'Сгенерировать новую ссылку',
        'list.makePrivate': 'Сделать приватным',
        'list.makePublic': 'Сделать публичным',
        'list.delete': 'Удалить список',
        'list.addWish': 'Добавить желание',
        'list.empty': 'В этом списке пока нет желаний',
        'list.emptyAction': 'Пока тут пусто, нажмите + чтобы добавить желание',
        'list.created': 'Список создан',
        'list.createFailed': 'Не удалось создать список',
        'list.createdAt': 'Создан',
        'list.updatedAt': 'Обновлён',
        'list.privateNotice': 'Доступ к этому списку предоставляется только по ссылке, не делитесь ей с посторонними!',
        'list.status.public': 'Публичный',
        'list.status.private': 'Приватный',
        'list.loadFailed': 'Не удалось загрузить список',
        'list.delete.confirmTitle': 'Удаление списка',
        'list.delete.confirmMessage': 'Вы уверены, что хотите удалить этот список?',
        'list.deleted': 'Список удалён',
        'list.deleteFailed': 'Не удалось удалить список',
        'list.privacy.changed': 'Приватность изменена!',
        'list.privacy.changeFailed': 'Не удалось изменить приватность',
        'list.share.missing': 'Не удалось получить ссылку',
        'list.share.copied': 'Ссылка скопирована',
        'list.share.copyFailed': 'Не удалось скопировать ссылку',
        'list.share.rotateConfirmTitle': 'Генерация новой ссылки',
        'list.share.rotateConfirmMessage': 'Старая ссылка перестанет работать. Продолжить?',
        'list.share.rotated': 'Новая ссылка сгенерирована!',
        'list.share.rotateFailed': 'Не удалось сгенерировать ссылку',
        'list.title.empty': 'Название не может быть пустым',
        'list.title.saveFailed': 'Не удалось сохранить название',

        // Wish detail
        'wish.title': 'Название',
        'wish.note': 'Описание',
        'wish.price': 'Стоимость',
        'wish.link': 'Ссылка',
        'wish.image': 'Изображение',
        'wish.selectImage': 'Нажмите для выбора изображения',
        'wish.delete': 'Удалить',
        'wish.save': 'Сохранить',
        'wish.cancel': 'Отменить',
        'wish.modal.title': 'Детали желания',
        'wish.createdAt': 'Создано',
        'wish.updatedAt': 'Изменено',
        'wish.sort.newFirst': 'Сначала новые',
        'wish.sort.oldFirst': 'Сначала старые',
        'wish.sort.nameAsc': 'По имени (А-Я)',
        'wish.sort.nameDesc': 'По имени (Я-А)',
        'wish.sort.updatedDesc': 'По обновлению (новые)',
        'wish.sort.updatedAsc': 'По обновлению (старые)',
        'wish.sort.priceDesc': 'По цене (дороже)',
        'wish.sort.priceAsc': 'По цене (дешевле)',
        'wish.filter.all': 'Все желания',
        'wish.filter.withImage': 'С картинкой',
        'wish.filter.withoutImage': 'Без картинки',
        'wish.filter.withNotes': 'С заметками',
        'wish.filter.withoutNotes': 'Без заметок',
        'wish.validation.titleRequired': 'Название обязательно',
        'wish.validation.priceIntegerNonNegative': 'Цена должна быть целым неотрицательным числом',
        'wish.updated': 'Желание обновлено',
        'wish.updateFieldFailed': 'Не удалось обновить поле',
        'wish.loadFailed': 'Не удалось загрузить желание',
        'wish.delete.confirmTitle': 'Удаление желания',
        'wish.delete.confirmMessage': 'Вы уверены, что хотите удалить это желание?',
        'wish.deleteFailed': 'Не удалось удалить желание',
        'wish.created': 'Желание создано',
        'wish.createFailed': 'Не удалось создать желание',
        'wish.createdImageUploadFailed': 'Желание создано, но изображение не загрузилось',
        'wish.reserve.action': 'Забронировать',
        'wish.reserve.release': 'Снять бронь',
        'wish.reserve.taken': 'Кем-то забронировано',
        'wish.reserve.success': 'Желание забронировано',
        'wish.reserve.released': 'Бронь снята',
        'wish.reserve.failed': 'Не удалось забронировать желание',
        'wish.reserve.releaseFailed': 'Не удалось снять бронь',
        'wish.image.updated': 'Изображение обновлено',
        'wish.image.updateFailed': 'Не удалось загрузить изображение',
        'wish.image.deleteTitle': 'Удаление изображения',
        'wish.image.deleteConfirm': 'Вы уверены, что хотите удалить изображение?',
        'wish.image.deleted': 'Изображение удалено',
        'wish.image.deleteFailed': 'Не удалось удалить изображение',
        'wish.placeholder.title': 'Без названия',
        'wish.placeholder.note': 'Нет заметок',
        'wish.placeholder.price': 'Цена не указана',
        'wish.placeholder.link': 'Ссылка не указана',

        // Theme labels
        'theme.light': 'Светлая',
        'theme.system': 'Системная',
        'theme.dark': 'Тёмная',

        // Profile
        'profile.label.name': 'Имя',
        'profile.label.username': 'Юзернейм',
        'profile.label.email': 'Электронная почта',
        'profile.changePassword': 'Изменить пароль',
        'profile.passwordModalTitle': 'Изменить пароль',
        'profile.password.oldPlaceholder': 'Старый пароль',
        'profile.password.newPlaceholder': 'Новый пароль',
        'profile.updated': 'Профиль обновлён',
        'profile.updateFailed': 'Не удалось обновить профиль',
        'profile.loadFailed': 'Не удалось загрузить профиль',
        'profile.avatar.updateSuccess': 'Аватар обновлён',
        'profile.avatar.uploadFailed': 'Не удалось загрузить аватар',
        'profile.avatar.removeTitle': 'Удаление аватара',
        'profile.avatar.removeConfirm': 'Удалить текущий аватар?',
        'profile.avatar.removeFailed': 'Не удалось удалить аватар',
        'profile.avatar.removed': 'Аватар удалён',
        'profile.password.changeSuccess': 'Пароль успешно изменен',
        'profile.password.changeFailed': 'Не удалось изменить пароль',

        // Auth
        'auth.logoutTitle': 'Выход',
        'auth.logoutConfirm': 'Вы уверены, что хотите выйти?',
        'auth.logoutFailed': 'Ошибка при выходе',
        'auth.loginError': 'Ошибка входа',
        'auth.invalidCredentials': 'Неверный юзернейм или пароль',
        'auth.loginRequestFailed': 'Ошибка при входе',
        'auth.registerError': 'Ошибка регистрации',
        'auth.registerRequestFailed': 'Ошибка при регистрации',

        // API errors
        'api.invalidRequestPayload': 'Некорректный запрос',
        'api.invalidJsonPayload': 'Некорректный JSON',
        'api.invalidCredentials': 'Неверный юзернейм или пароль',
        'api.nameRequired': 'Имя обязательно',
        'api.usernameRequired': 'Юзернейм обязателен',
        'api.emailRequired': 'Почта обязательна',
        'api.emailInvalid': 'Введите корректную почту',
        'api.passwordRequired': 'Пароль обязателен',
        'api.passwordMin': 'Пароль должен быть не короче 8 символов',
        'api.usernameCharset': 'Юзернейм может содержать только латинские буквы, русские буквы, цифры, _ и -',
        'api.newPasswordRequired': 'Новый пароль обязателен',
        'api.newPasswordMin': 'Новый пароль должен быть не короче 8 символов',
        'api.currentPasswordRequired': 'Текущий пароль обязателен',
        'api.refreshTokenRequired': 'Refresh token обязателен',
        'api.tokenRequired': 'Токен обязателен',
        'api.usernameTaken': 'Этот юзернейм уже занят',
        'api.emailInUse': 'Эта почта уже используется',
        'api.userNotFound': 'Пользователь не найден',
        'api.listNotFound': 'Список не найден',
        'api.wishNotFound': 'Желание не найдено',
        'api.invalidResetToken': 'Ссылка для сброса пароля недействительна или устарела',
        'api.invalidVerificationToken': 'Ссылка подтверждения недействительна или устарела',
        'api.wrongPassword': 'Неверный пароль',
        'api.wrongCurrentPassword': 'Неверный текущий пароль',
        'api.wishAlreadyReserved': 'Желание уже забронировано',
        'api.wishNotReservedByYou': 'Это желание забронировано не вами',
        'api.cannotReserveOwnWish': 'Нельзя бронировать собственное желание',
        'api.wishWrongList': 'Желание не принадлежит этому списку',
        'api.listPrivate': 'Этот вишлист приватный',
        'api.notWishlistOwner': 'Вы не владелец этого вишлиста',
        'api.notWishOwner': 'Вы не владелец этого желания',
        'api.searchQueryTooShort': 'Введите хотя бы 2 символа',

        // Dialogs
        'dialog.alertTitle': 'Уведомление',
        'dialog.confirmTitle': 'Подтверждение',
        'dialog.promptTitle': 'Ввод',
        'dialog.confirm': 'Подтвердить',

        // Forms
        'form.unsaved.title': 'Несохраненные изменения',
        'form.unsaved.message': 'У вас есть несохраненные изменения. Закрыть без сохранения?',

        // 404 page
        'notFound.message': 'Страница не найдена',
        'notFound.backHome': 'Вернуться на главную',

        // Common
        'common.close': 'Закрыть',
        'common.save': 'Сохранить',
        'common.cancel': 'Отменить',
        'common.ok': 'OK',
        'common.delete': 'Удалить',
        'common.edit': 'Редактировать',
        'common.loading': 'Загрузка...',
        'common.notSpecified': 'Не указано',
        'common.error': 'Ошибка',
        'common.success': 'Успешно',
    },
    en: {
        // Header
        'header.login': 'Get started',
        'header.myLists': 'My wishlists',
        'header.profile': 'Profile',
        'header.logout': 'Logout',

        // Home page
        'home.title': 'Wishlist',
        'home.subtitle': 'Create wishlists and share them with loved ones',
        'home.feature1.title': 'Organize',
        'home.feature1.desc': 'Create lists for any occasions – birthdays, holidays, dreams',
        'home.feature2.title': 'Control',
        'home.feature2.desc': 'Public or private lists with access by shared link',
        'home.feature3.title': 'Plan',
        'home.feature3.desc': 'See price history and get notified about discounts',
        'home.feature3.badge': 'Someday...',

        // Auth modal
        'auth.login': 'Login',
        'auth.register': 'Register',
        'auth.username': 'Username',
        'auth.password': 'Password',
        'auth.email': 'Email',
        'auth.name': 'Name',
        'auth.loginButton': 'Log in',
        'auth.registerButton': 'Sign up',
        'auth.forgotPassword': 'Forgot password?',
        'auth.forgotPasswordTitle': 'Reset password',
        'auth.forgotPasswordHint': 'Enter your account email and we will send you a reset link',
        'auth.forgotPasswordSubmit': 'Send reset link',
        'auth.forgotPasswordSuccess': 'If an account with this email exists, a password reset email has been sent',
        'auth.forgotPasswordError': 'Failed to send password reset email',
        'auth.resetPasswordTitle': 'New password',
        'auth.resetPasswordHint': 'Enter a new password for your account',
        'auth.resetPasswordSubmit': 'Save new password',
        'auth.resetPasswordNew': 'New password',
        'auth.resetPasswordConfirm': 'Repeat password',
        'auth.resetPasswordSuccess': 'Password changed successfully',
        'auth.resetPasswordError': 'Failed to change password',
        'auth.resetPasswordMismatch': 'Passwords do not match',
        'auth.resetPasswordInvalidToken': 'Password reset link is invalid or expired',

        // Lists page
        'lists.title': 'My wishlists',
        'lists.search': 'Search...',
        'lists.sort': 'Sort',
        'lists.filter': 'Filter',
        'lists.create': 'Create list',
        'lists.empty': 'You don\'t have any lists yet',
        'lists.emptyTitle': 'You have no lists yet',
        'lists.emptySubtitle': 'Create your first wishlist',
        'lists.sort.newFirst': 'Newest first',
        'lists.sort.oldFirst': 'Oldest first',
        'lists.sort.nameAsc': 'Name (A-Z)',
        'lists.sort.nameDesc': 'Name (Z-A)',
        'lists.sort.updatedDesc': 'Updated (newest)',
        'lists.sort.updatedAsc': 'Updated (oldest)',
        'lists.sort.wishesDesc': 'Wish count (high to low)',
        'lists.sort.wishesAsc': 'Wish count (low to high)',
        'lists.filter.all': 'All lists',
        'lists.filter.public': 'Public',
        'lists.filter.private': 'Private',
        'lists.filter.empty': 'Empty',
        'lists.filter.notEmpty': 'Not empty',
        'lists.loadFailed': 'Failed to load lists',
        'lists.back': 'Back',
        'users.find.title': 'Find user',
        'users.find.placeholder': 'Enter username',
        'users.find.submit': 'Open public wishlists',
        'users.find.notFound': 'User not found',
        'users.find.failed': 'No such user was found',
        'users.find.searchEmpty': 'Start typing a username',
        'users.find.searchNoResults': 'No users found',
        'users.find.searchChoose': 'Choose a user from the list',
        'users.publicLists.back': 'Back to my wishlists',
        'users.publicLists.subtitle': 'Only public wishlists are shown here',
        'users.publicLists.emptyTitle': 'Nothing here',
        'users.publicLists.emptySubtitle': 'Looks like this user has no wishlists yet',

        // List detail
        'list.copyLink': 'Copy link',
        'list.generateLink': 'Generate new link',
        'list.makePrivate': 'Make private',
        'list.makePublic': 'Make public',
        'list.delete': 'Delete list',
        'list.addWish': 'Add wish',
        'list.empty': 'This list has no wishes yet',
        'list.emptyAction': 'List is empty for now, press + to add a wish',
        'list.created': 'List created',
        'list.createFailed': 'Failed to create list',
        'list.createdAt': 'Created',
        'list.updatedAt': 'Updated',
        'list.privateNotice': 'Access to this wishlist is granted only by link, do not share it with strangers!',
        'list.status.public': 'Public',
        'list.status.private': 'Private',
        'list.loadFailed': 'Failed to load list',
        'list.delete.confirmTitle': 'Delete list',
        'list.delete.confirmMessage': 'Are you sure you want to delete this list?',
        'list.deleted': 'List deleted',
        'list.deleteFailed': 'Failed to delete list',
        'list.privacy.changed': 'Privacy updated!',
        'list.privacy.changeFailed': 'Failed to change privacy',
        'list.share.missing': 'Failed to get link',
        'list.share.copied': 'Link copied',
        'list.share.copyFailed': 'Failed to copy link',
        'list.share.rotateConfirmTitle': 'Generate new link',
        'list.share.rotateConfirmMessage': 'The old link will stop working. Continue?',
        'list.share.rotated': 'New link generated!',
        'list.share.rotateFailed': 'Failed to generate new link',
        'list.title.empty': 'Title cannot be empty',
        'list.title.saveFailed': 'Failed to save title',

        // Wish detail
        'wish.title': 'Title',
        'wish.note': 'Note',
        'wish.price': 'Price',
        'wish.link': 'Link',
        'wish.image': 'Image',
        'wish.selectImage': 'Click to select an image',
        'wish.delete': 'Delete',
        'wish.save': 'Save',
        'wish.cancel': 'Cancel',
        'wish.modal.title': 'Wish details',
        'wish.createdAt': 'Created',
        'wish.updatedAt': 'Updated',
        'wish.sort.newFirst': 'Newest first',
        'wish.sort.oldFirst': 'Oldest first',
        'wish.sort.nameAsc': 'Name (A-Z)',
        'wish.sort.nameDesc': 'Name (Z-A)',
        'wish.sort.updatedDesc': 'Updated (newest)',
        'wish.sort.updatedAsc': 'Updated (oldest)',
        'wish.sort.priceDesc': 'Price (high to low)',
        'wish.sort.priceAsc': 'Price (low to high)',
        'wish.filter.all': 'All wishes',
        'wish.filter.withImage': 'With image',
        'wish.filter.withoutImage': 'Without image',
        'wish.filter.withNotes': 'With notes',
        'wish.filter.withoutNotes': 'Without notes',
        'wish.validation.titleRequired': 'Title is required',
        'wish.validation.priceIntegerNonNegative': 'Price must be a non-negative integer',
        'wish.updated': 'Wish updated',
        'wish.updateFieldFailed': 'Failed to update field',
        'wish.loadFailed': 'Failed to load wish',
        'wish.delete.confirmTitle': 'Delete wish',
        'wish.delete.confirmMessage': 'Are you sure you want to delete this wish?',
        'wish.deleteFailed': 'Failed to delete wish',
        'wish.created': 'Wish created',
        'wish.createFailed': 'Failed to create wish',
        'wish.createdImageUploadFailed': 'Wish created, but image upload failed',
        'wish.reserve.action': 'Reserve',
        'wish.reserve.release': 'Release reservation',
        'wish.reserve.taken': 'Reserved by someone',
        'wish.reserve.success': 'Wish reserved',
        'wish.reserve.released': 'Reservation removed',
        'wish.reserve.failed': 'Failed to reserve wish',
        'wish.reserve.releaseFailed': 'Failed to release reservation',
        'wish.image.updated': 'Image updated',
        'wish.image.updateFailed': 'Failed to upload image',
        'wish.image.deleteTitle': 'Delete image',
        'wish.image.deleteConfirm': 'Are you sure you want to delete the image?',
        'wish.image.deleted': 'Image deleted',
        'wish.image.deleteFailed': 'Failed to delete image',
        'wish.placeholder.title': 'No title',
        'wish.placeholder.note': 'No notes',
        'wish.placeholder.price': 'No price',
        'wish.placeholder.link': 'No link',

        // Theme labels
        'theme.light': 'Light',
        'theme.system': 'System',
        'theme.dark': 'Dark',

        // Profile
        'profile.label.name': 'Name',
        'profile.label.username': 'Username',
        'profile.label.email': 'Email',
        'profile.changePassword': 'Change password',
        'profile.passwordModalTitle': 'Change password',
        'profile.password.oldPlaceholder': 'Old password',
        'profile.password.newPlaceholder': 'New password',
        'profile.updated': 'Profile updated',
        'profile.updateFailed': 'Failed to update profile',
        'profile.loadFailed': 'Failed to load profile',
        'profile.avatar.updateSuccess': 'Avatar updated',
        'profile.avatar.uploadFailed': 'Failed to upload avatar',
        'profile.avatar.removeTitle': 'Remove avatar',
        'profile.avatar.removeConfirm': 'Delete current avatar?',
        'profile.avatar.removeFailed': 'Failed to remove avatar',
        'profile.avatar.removed': 'Avatar removed',
        'profile.password.changeSuccess': 'Password changed successfully',
        'profile.password.changeFailed': 'Failed to change password',

        // Auth
        'auth.logoutTitle': 'Logout',
        'auth.logoutConfirm': 'Are you sure you want to log out?',
        'auth.logoutFailed': 'Logout failed',
        'auth.loginError': 'Login error',
        'auth.invalidCredentials': 'Invalid username or password',
        'auth.loginRequestFailed': 'Login failed',
        'auth.registerError': 'Registration error',
        'auth.registerRequestFailed': 'Registration failed',

        // API errors
        'api.invalidRequestPayload': 'Invalid request payload',
        'api.invalidJsonPayload': 'Invalid JSON payload',
        'api.invalidCredentials': 'Invalid username or password',
        'api.nameRequired': 'Name is required',
        'api.usernameRequired': 'Username is required',
        'api.emailRequired': 'Email is required',
        'api.emailInvalid': 'Enter a valid email address',
        'api.passwordRequired': 'Password is required',
        'api.passwordMin': 'Password must be at least 8 characters',
        'api.usernameCharset': 'Username may contain only latin letters, russian letters, digits, underscore, and hyphen',
        'api.newPasswordRequired': 'New password is required',
        'api.newPasswordMin': 'New password must be at least 8 characters',
        'api.currentPasswordRequired': 'Current password is required',
        'api.refreshTokenRequired': 'Refresh token is required',
        'api.tokenRequired': 'Token is required',
        'api.usernameTaken': 'This username is already taken',
        'api.emailInUse': 'This email is already in use',
        'api.userNotFound': 'User not found',
        'api.listNotFound': 'List not found',
        'api.wishNotFound': 'Wish not found',
        'api.invalidResetToken': 'Password reset link is invalid or expired',
        'api.invalidVerificationToken': 'Verification link is invalid or expired',
        'api.wrongPassword': 'Wrong password',
        'api.wrongCurrentPassword': 'Wrong current password',
        'api.wishAlreadyReserved': 'Wish is already reserved',
        'api.wishNotReservedByYou': 'Wish is not reserved by you',
        'api.cannotReserveOwnWish': 'You cannot reserve your own wish',
        'api.wishWrongList': 'Wish does not belong to this list',
        'api.listPrivate': 'This wishlist is private',
        'api.notWishlistOwner': 'You are not the owner of this wishlist',
        'api.notWishOwner': 'You are not the owner of this wish',
        'api.searchQueryTooShort': 'Enter at least 2 characters',

        // Dialogs
        'dialog.alertTitle': 'Notification',
        'dialog.confirmTitle': 'Confirmation',
        'dialog.promptTitle': 'Input',
        'dialog.confirm': 'Confirm',

        // Forms
        'form.unsaved.title': 'Unsaved changes',
        'form.unsaved.message': 'You have unsaved changes. Close without saving?',

        // 404 page
        'notFound.message': 'Page not found',
        'notFound.backHome': 'Back to home',

        // Common
        'common.close': 'Close',
        'common.save': 'Save',
        'common.cancel': 'Cancel',
        'common.ok': 'OK',
        'common.delete': 'Delete',
        'common.edit': 'Edit',
        'common.loading': 'Loading...',
        'common.notSpecified': 'Not specified',
        'common.error': 'Error',
        'common.success': 'Success',
    }
};

// Current language (default: ru)
let currentLang = localStorage.getItem('wishlist_lang') || 'ru';

// Get translation by key
function t(key) {
    return translations[currentLang][key] || key;
}

const apiErrorTranslationMap = {
    'invalid request payload': 'api.invalidRequestPayload',
    'invalid json payload': 'api.invalidJsonPayload',
    'invalid credentials': 'api.invalidCredentials',
    'name is required': 'api.nameRequired',
    'username is required': 'api.usernameRequired',
    'email is required': 'api.emailRequired',
    'email must be a valid email address': 'api.emailInvalid',
    'password is required': 'api.passwordRequired',
    'password must be at least 8 characters': 'api.passwordMin',
    'username may contain only latin letters, russian letters, digits, underscore, and hyphen': 'api.usernameCharset',
    'new password is required': 'api.newPasswordRequired',
    'new password must be at least 8 characters': 'api.newPasswordMin',
    'current password is required': 'api.currentPasswordRequired',
    'refresh token is required': 'api.refreshTokenRequired',
    'token is required': 'api.tokenRequired',
    'username is already taken': 'api.usernameTaken',
    'email is already in use': 'api.emailInUse',
    'user not found': 'api.userNotFound',
    'list not found': 'api.listNotFound',
    'wish with id not found': 'api.wishNotFound',
    'invalid or expired password reset token': 'api.invalidResetToken',
    'invalid or expired verification token': 'api.invalidVerificationToken',
    'wrong password': 'api.wrongPassword',
    'wrong current password': 'api.wrongCurrentPassword',
    'wish is already reserved': 'api.wishAlreadyReserved',
    'wish is not reserved by you': 'api.wishNotReservedByYou',
    'you cannot reserve your own wish': 'api.cannotReserveOwnWish',
    'wish does not belong to this list': 'api.wishWrongList',
    'this wishlist is private': 'api.listPrivate',
    'you are not the owner of this wishlist': 'api.notWishlistOwner',
    'you are not the owner of this wish': 'api.notWishOwner',
    'search query too short': 'api.searchQueryTooShort',
};

function capitalizeMessage(message) {
    const value = (message || '').toString().trim();
    if (!value) return '';
    return value.charAt(0).toUpperCase() + value.slice(1);
}

function normalizeApiErrorLookupKey(message) {
    return (message || '')
        .toString()
        .trim()
        .toLowerCase()
        .replace(/\s+/g, ' ')
        .replace(/'[^']+'/g, '')
        .replace(/"[^"]+"/g, '');
}

function translateApiErrorMessage(message, fallback = '') {
    const lookupKey = normalizeApiErrorLookupKey(message);
    const translationKey = apiErrorTranslationMap[lookupKey];
    if (translationKey) {
        return t(translationKey);
    }

    if (lookupKey.startsWith('user with ') && lookupKey.endsWith(' not found')) {
        return t('api.userNotFound');
    }
    if (lookupKey.startsWith('list with ') && lookupKey.endsWith(' not found')) {
        return t('api.listNotFound');
    }
    if (lookupKey.startsWith('wish with ') && lookupKey.endsWith(' not found')) {
        return t('api.wishNotFound');
    }

    if (!message) {
        return fallback ? t(fallback) : '';
    }

    return capitalizeMessage(message);
}

function syncLanguageMenuSelection() {
    const menu = document.getElementById('langMenu');
    if (!menu) return;

    menu.querySelectorAll('.popup-menu-item').forEach((item) => {
        const onclick = item.getAttribute('onclick') || '';
        const isSelected = onclick.includes(`'${currentLang}'`) || onclick.includes(`"${currentLang}"`);
        item.classList.toggle('selected', isSelected);
    });
}

// Set language
function setLanguage(lang) {
    if (!translations[lang]) return; // Invalid language

    currentLang = lang; // Update current language
    localStorage.setItem('wishlist_lang', lang); // Save to localStorage
    updatePageTranslations(); // Update all text on page
    syncLanguageMenuSelection(); // Sync checkmark state in FAB menu
}

// Update all translatable elements on the page
function updatePageTranslations() {
    // Find all elements with data-i18n attribute
    document.querySelectorAll('[data-i18n]').forEach(element => {
        const key = element.getAttribute('data-i18n'); // Get translation key
        element.textContent = t(key); // Set translated text
    });

    // Find all elements with data-i18n-placeholder attribute
    document.querySelectorAll('[data-i18n-placeholder]').forEach(element => {
        const key = element.getAttribute('data-i18n-placeholder'); // Get translation key
        element.placeholder = t(key); // Set translated placeholder
    });
}

// Toggle between languages
function toggleLanguage() {
    const newLang = currentLang === 'ru' ? 'en' : 'ru'; // Switch language
    setLanguage(newLang); // Apply new language
}

// Initialize i18n on page load
document.addEventListener('DOMContentLoaded', () => {
    updatePageTranslations(); // Translate all elements
    syncLanguageMenuSelection(); // Sync checkmark state in FAB menu
});
