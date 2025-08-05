package obfuscator

// Общие определения для Anti-VM/Anti-Debug, доступны на всех платформах.

// vmCheckVarName — уникальное имя глобальной переменной результата VM-проверки,
// используется в Anti-VM и Anti-Debug для «weaving» ключей и условий.
var vmCheckVarName = NewName()
