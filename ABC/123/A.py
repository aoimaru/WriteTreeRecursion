# coding: utf-8
# Your code here!

def main():
    N, Q = map(int, input().split())
    S = input()
    cum = [0]*(N+1)
    for i in range(N-1):
        if (S[i] == "A" and S[i+1] == "C"):
            cum[i+1] = cum[i]+1
        else:
            cum[i+1] = cum[i]

    for _ in range(Q):
        L, R = map(int, input().split())
        print(cum[R-1]-cum[L-1])

if __name__ == "__main__":
    main()