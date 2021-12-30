#ifndef __LOAD_CONST_H
#define __LOAD_CONST_H

#define LOAD_OFFSET(param, var) asm("%0 = " param " ll" : "=r"(var))

#endif // !__LOAD_CONST_H
